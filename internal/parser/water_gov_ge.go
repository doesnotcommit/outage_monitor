package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/doesnotcommit/outage_monitor/internal/outage"
)

type WaterGovGe struct {
	c                         *http.Client
	outageDateTimeLayout      string
	mapMarkersRx              *regexp.Regexp
	outageStartLabelRx        *regexp.Regexp
	outageEndLabelRx          *regexp.Regexp
	outageAffectedCustomersRx *regexp.Regexp
	outageAddressesRx         *regexp.Regexp
	location                  *time.Location
	mapURI                    string
	problemURITpl             string
	sl                        *slog.Logger
}

type waterGovGePoint struct {
	Problem bool   `json:"problem"`
	Id      string `json:"id"`
	Title   string `json:"title"`
	Lat     string `json:"lat"`
	Lng     string `json:"lng"`
}

func NewWaterGovGe(sl *slog.Logger) (WaterGovGe, error) {
	const (
		mapURI               = "http://water.gov.ge/page/map"
		problemURITpl        = "http://water.gov.ge/page/problem/%s"
		outageDateTimeLayout = "02/01/2006 15:04:05"
	)
	tbilisi, err := time.LoadLocation("Asia/Tbilisi")
	if err != nil {
		return WaterGovGe{}, fmt.Errorf("load location: %w", err)
	}
	var (
		mapMarkersRx              = regexp.MustCompile(`var markers = (\[.+\]);`)
		outageStartLabelRx        = regexp.MustCompile(`<div>\s+წყალმომარაგების\s+შეწყვეტის\s+დრო:\s+([\d/\s:]+)\s+</div>`)
		outageEndLabelRx          = regexp.MustCompile(`<div>\s+წყალმომარაგების\s+აღდგენის\s+დრო:\s+([\d/\s:]+)\s+</div>`)
		outageAffectedCustomersRx = regexp.MustCompile(`<div>\sგამორთული\sაბონენტების\sრაოდენობა:\s(\d+)\s</div>`)
		outageAddressesRx         = regexp.MustCompile(`<div>\s*([^<>]+)\s*</div>`)
	)
	c := http.Client{
		Timeout: time.Second * 10,
	}
	return WaterGovGe{
		&c,
		outageDateTimeLayout,
		mapMarkersRx,
		outageStartLabelRx,
		outageEndLabelRx,
		outageAffectedCustomersRx,
		outageAddressesRx,
		tbilisi,
		mapURI,
		problemURITpl,
		sl,
	}, nil
}

func (w WaterGovGe) GetOutages(ctx context.Context) ([]outage.WaterGovGe, error) {
	handleErr := func(err error) ([]outage.WaterGovGe, error) {
		return nil, fmt.Errorf("get outages: %w", err)
	}
	rawMapHTML, err := w.fetchRawHTMLFile(ctx, w.mapURI)
	if err != nil {
		return handleErr(err)
	}
	points, err := w.parseMapMarkers(ctx, rawMapHTML)
	if err != nil {
		return handleErr(err)
	}
	problems, err := w.parseProblems(ctx, points)
	if err != nil {
		return handleErr(err)
	}
	return problems, nil
}

func (w WaterGovGe) parseProblems(ctx context.Context, points []waterGovGePoint) ([]outage.WaterGovGe, error) {
	handleErr := func(err error) ([]outage.WaterGovGe, error) {
		return nil, fmt.Errorf("parse problems: %w", err)
	}
	var problems []outage.WaterGovGe
	for _, point := range points {
		if !point.Problem {
			continue
		}
		rawProblemHTML, err := w.fetchRawHTMLFile(ctx, fmt.Sprintf(w.problemURITpl, point.Id))
		if err != nil {
			return handleErr(err)
		}
		shortTitleGe, foundSfx := strings.CutSuffix(strings.TrimSpace(point.Title), " სერვის ცენტრი")
		if !foundSfx {
			w.sl.Warn("no suffix found")
		}
		location := outage.Location{
			Id:       strings.TrimSpace(point.Id),
			TitleGe:  shortTitleGe,
			TitleLat: translit(shortTitleGe),
			Lat:      strings.TrimSpace(point.Lat),
			Lng:      strings.TrimSpace(point.Lng),
		}
		problem, err := w.parseProblem(ctx, location, rawProblemHTML)
		if err != nil {
			return handleErr(err)
		}
		problems = append(problems, problem)
	}
	return problems, nil
}

func (w WaterGovGe) parseProblem(ctx context.Context, location outage.Location, rawProblemHTML []byte) (outage.WaterGovGe, error) {
	handleErr := func(err error) (outage.WaterGovGe, error) {
		return outage.WaterGovGe{}, fmt.Errorf("parse problem [%s]: %w", string(rawProblemHTML), err)
	}
	rawOutageStartBytes := w.outageStartLabelRx.FindSubmatch(rawProblemHTML)
	if len(rawOutageStartBytes) < 2 {
		return handleErr(errNoOutageStart)
	}
	rawOutageStart := strings.TrimSpace(string(rawOutageStartBytes[1]))
	rawOutageEndBytes := w.outageEndLabelRx.FindSubmatch(rawProblemHTML)
	if len(rawOutageStartBytes) < 2 {
		return handleErr(errNoOutageEnd)
	}
	rawOutageEnd := strings.TrimSpace(string(rawOutageEndBytes[1]))
	outageStart, err := time.ParseInLocation(w.outageDateTimeLayout, rawOutageStart, w.location)
	if err != nil {
		return handleErr(err)
	}
	outageEnd, err := time.ParseInLocation(w.outageDateTimeLayout, rawOutageEnd, w.location)
	if err != nil {
		return handleErr(err)
	}
	rawOutageAffectedCustomersBytes := w.outageAffectedCustomersRx.FindSubmatch(rawProblemHTML)
	if len(rawOutageStartBytes) < 2 {
		return handleErr(errNoOutageStart)
	}
	outageAffectedCustomersStr := strings.TrimSpace(string(rawOutageAffectedCustomersBytes[1]))
	outageAffectedCustomers, err := strconv.Atoi(outageAffectedCustomersStr)
	if err != nil {
		return handleErr(err)
	}
	rawAddressesBytes := w.outageAddressesRx.FindAllSubmatch(rawProblemHTML, -1)
	if len(rawAddressesBytes) < 4 {
		return handleErr(errNoAddresses)
	}
	rawAddressesBytes = rawAddressesBytes[3:]
	var addresses []string
	for _, rawAddr := range rawAddressesBytes {
		if len(rawAddr) < 2 {
			continue
		}
		addr := strings.TrimSpace(html.UnescapeString(string(rawAddr[1])))
		addresses = append(addresses, addr)
	}
	return outage.WaterGovGe{
		Start:             outageStart,
		End:               outageEnd,
		AffectedCustomers: outageAffectedCustomers,
		AddressesGe:       addresses,
		Location:          location,
	}, nil
}

func (w WaterGovGe) parseMapMarkers(ctx context.Context, htmlFile []byte) ([]waterGovGePoint, error) {
	handleErr := func(err error) ([]waterGovGePoint, error) {
		return nil, fmt.Errorf("parse map markers [%s]: %w", string(htmlFile), err)
	}
	submatch := w.mapMarkersRx.FindSubmatch(htmlFile)
	if len(submatch) < 2 {
		return handleErr(errMapNotFound)
	}
	var points []waterGovGePoint
	if err := json.Unmarshal(submatch[1], &points); err != nil {
		return handleErr(err)
	}
	return points, nil
}

func (w WaterGovGe) fetchRawHTMLFile(ctx context.Context, addr string) ([]byte, error) {
	handleErr := func(err error) ([]byte, error) {
		return nil, fmt.Errorf("fetch html file at %s: %w", addr, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return handleErr(err)
	}
	resp, err := w.c.Do(req)
	if err != nil {
		return handleErr(err)
	}
	if resp.Body == nil {
		return handleErr(errNoRespBody)
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return handleErr(err)
	}
	return rawBody, nil
}

func translit(ge string) string {
	table := map[rune]string{
		'ა': "a", 'ბ': "b", 'გ': "g",
		'დ': "d", 'ე': "e", 'ვ': "v",
		'ზ': "z", 'თ': "t", 'ი': "i",
		'კ': "k'", 'ლ': "l", 'მ': "m",
		'ნ': "n", 'ო': "o", 'პ': "p'",
		'ჟ': "zh", 'რ': "r", 'ს': "s",
		'ტ': "t'", 'უ': "u", 'ფ': "p",
		'ქ': "k", 'ღ': "gh", 'ყ': "q",
		'შ': "sh", 'ჩ': "ch", 'ც': "ts",
		'ძ': "dz", 'წ': "ts'", 'ჭ': "ch'",
		'ხ': "kh", 'ჯ': "j", 'ჰ': "h",
	}
	result := make([]rune, 0, len(ge))
	for _, r := range ge {
		tr, ok := table[r]
		if !ok {
			tr = " "
		}
		result = append(result, []rune(tr)...)
	}
	return string(result)
}
