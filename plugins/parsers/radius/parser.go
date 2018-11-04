package radius

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type RadiusInfo map[string]string

type RadiusParser struct {
	MetricName      string
	MeasurementName string
	TagKeys         []string
	TimestampColumn string
	TimestampFormat string
	DefaultTags     map[string]string
	TimeFunc        func() time.Time
	_info           RadiusInfo
}

func (p *RadiusParser) SetTimeFunc(fn metric.TimeFunc) {
	p.TimeFunc = fn
}

func (p *RadiusParser) compile(r *bytes.Reader) (*bufio.Scanner, error) {
	scanner := bufio.NewScanner(r)
	return scanner, nil
}

func (p *RadiusParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	info := RadiusInfo{}

	r := bytes.NewReader(buf)
	scanner, err := p.compile(r)
	if err != nil {
		return nil, err
	}
	metrics := make([]telegraf.Metric, 0)
	for scanner.Scan() {
		line := scanner.Text()
		t, err := time.Parse(p.TimestampFormat, line)
		if err == nil {
			info["Timestamp"] = strconv.FormatInt(t.Unix(), 10)
		}
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = trimQuotes(strings.TrimSpace(line[equal+1:]))
				}
				info[key] = value
			}
		}
		if len(line) == 0 {
			m, err := p.parseRecord(info)
			if err != nil {
				return metrics, err
			}
			metrics = append(metrics, m)
			for k, _ := range info {
				delete(info, k)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return metrics, err
	}
	return metrics, nil
}

func (p *RadiusParser) ParseLine(line string) (telegraf.Metric, error) {
	t, err := time.Parse(p.TimestampFormat, line)
	if p._info == nil {
		p._info = RadiusInfo{}
	}
	if err == nil {
		p._info["Timestamp"] = strconv.FormatInt(t.Unix(), 10)
	}
	if equal := strings.Index(line, "="); equal >= 0 {
		if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
			value := ""
			if len(line) > equal {
				value = trimQuotes(strings.TrimSpace(line[equal+1:]))
			}
			p._info[key] = value
		}
	}
	if len(line) == 0 {
		m, err := p.parseRecord(p._info)
		if err != nil {
			return m, err
		}
		for k, _ := range p._info {
			delete(p._info, k)
		}

		return m, nil
	}

	return nil, nil
}

func (p *RadiusParser) parseRecord(info RadiusInfo) (telegraf.Metric, error) {
	record := make(map[string]interface{})
	tags := make(map[string]string)

	// add default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}
	for fieldName, value := range info {
		// attempt type conversions
		if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			record[fieldName] = iValue
		} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
			record[fieldName] = fValue
		} else if bValue, err := strconv.ParseBool(value); err == nil {
			record[fieldName] = bValue
		} else {
			record[fieldName] = value
		}
	}

	// will default to plugin name
	measurementName := p.MetricName
	if p.MeasurementName != "" {
		measurementName = p.MeasurementName
	}
	metricTime := p.TimeFunc()
	if p.TimestampColumn != "" {
		if record[p.TimestampColumn] == nil {
			return nil, fmt.Errorf("timestamp column: %v could not be found", p.TimestampColumn)
		}
		tStr := fmt.Sprintf("%v", record[p.TimestampColumn])
		if p.TimestampFormat == "" {
			return nil, fmt.Errorf("timestamp format must be specified")
		}

		i, err := strconv.ParseInt(tStr, 10, 64)
		if err == nil {
			metricTime = time.Unix(i, 0)
		} else {
			metricTime, err = time.Parse(p.TimestampFormat, tStr)
			if err != nil {
				return nil, err
			}
		}
	}

	m, err := metric.New(measurementName, tags, record, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *RadiusParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func trimQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
