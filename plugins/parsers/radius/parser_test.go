package radius

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

func TestBasic(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	User-Name = "imgtrunk"
	NAS-IP-Address = 172.16.31.4
	NAS-Port = 1813
	Calling-Station-Id = "33139585858"
	Called-Station-Id = "33608635117"
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Start
	NAS-Port-Type = Ethernet
	Service-Type = Login-User
	Dialogic-call-origin = "originate"
	Dialogic-call-type = "SIP"
	Acct-Delay-Time = 0
	Login-IP-Host = 172.16.31.13
	Tunnel-Client-Endpoint:0 = "172.16.31.4"
	Dialogic-setup-time = "SUN SEP 16 22:02:21:236 2018"
	Dialogic-voip-dst-sig-ip-in = "79.170.216.134"
	Dialogic-voip-dst-rtp-ip-in = "79.170.216.134"
	Dialogic-dnis-pre-translate = "608635117"
	Dialogic-ani-pre-translate = "33139585858"
	Dialogic-call-direction = "INCOMING LEG"
	Dialogic-trunk-grp-in = "NERIM_SBC_Peering"
	Dialogic-voip-src-rtp-ip-in = "79.170.216.52"
	Dialogic-voip-src-sig-ip-in = "79.170.216.50"
	Dialogic-call-id = "0705bd41-349f-1237-248b-ad578c8ed49b"
	Dialogic-prev-hop-ip = "79.170.216.134:5062/UDP"
	Dialogic-prev-hop-via = "sip:79.170.216.134:5062"
	Dialogic-incoming-req-uri = "sip:33608635117@img-sde-1"
	Dialogic-voip-local-vocoders = "PCMA,PCMU,telephone-event"
	Dialogic-voip-remote-vocoders = "PCMA,PCMU,telephone-event"
	Dialogic-voip-codec-priority = "Remote"
	Dialogic-Attr-154 = 0x34313337
	Dialogic-Attr-155 = 0x3236
	Client-IP-Address = 172.16.31.4
	Acct-Unique-Session-Id = "d6ae6cb422467ab8"
	Timestamp = 1537135341

`

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	require.Equal(t, "radius", metrics[0].Name())
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1537135341000000000))
}

func TestTimestampError(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	Calling-Station-Id = "33139585858"
	Called-Station-Id = "33608635117"
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Start
	NAS-Port-Type = Ethernet
	Service-Type = Login-User
	Client-IP-Address = 172.16.31.4
	Acct-Unique-Session-Id = "d6ae6cb422467ab8"

`

	_, err := p.Parse([]byte(data))
	require.Equal(t, fmt.Errorf("timestamp format must be specified"), err)
}

func TestArrayMetrics(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Start
	Timestamp = 1537135341

Mon Sep 17 00:02:22 2018
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Stop
	Timestamp = 1537135342

`

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "radius", metrics[0].Name())
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1537135341000000000))
	require.Equal(t, "Start", metrics[0].Fields()["Acct-Status-Type"])
	require.Equal(t, "radius", metrics[1].Name())
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1537135342000000000))
	require.Equal(t, "Stop", metrics[1].Fields()["Acct-Status-Type"])
}

func TestQuotedCharacter(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Unique-Session-Id = "d6ae6cb422467ab8"

`

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	require.Equal(t, "00201c14283a008f00841b9ed2ed55ca02c7", metrics[0].Fields()["Acct-Session-Id"])
	require.Equal(t, "d6ae6cb422467ab8", metrics[0].Fields()["Acct-Unique-Session-Id"])

}

func TestValueConversion(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	first = 3.3
	second = 4
	third = true
	fourth = "hello"

`

	expectedTags := make(map[string]string)
	expectedFields := map[string]interface{}{
		"first":     3.3,
		"second":    4,
		"third":     true,
		"fourth":    "hello",
		"Timestamp": 1537142541,
	}

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)

	expectedMetric, err1 := metric.New("test_value", expectedTags, expectedFields, time.Unix(0, 0))
	returnedMetric, err2 := metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))
	require.NoError(t, err1)
	require.NoError(t, err2)

	//deep equal fields
	require.Equal(t, expectedMetric.Fields(), returnedMetric.Fields())
}

func TestTrimSpace(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	data := `Mon Sep 17 00:02:21 2018
	first =       3.3   
	second = 4      
	third=         true
	fourth     =     "hello"   

`

	expectedFields := map[string]interface{}{
		"first":     3.3,
		"second":    int64(4),
		"third":     true,
		"fourth":    "hello",
		"Timestamp": int64(1537142541),
	}

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[0].Fields())
}

func TestParseStream(t *testing.T) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	var data = []string{"Mon Sep 17 00:02:21 2018",
		"Acct-Session-Id = \"00201c14283a008f00841b9ed2ed55ca02c7\"",
		"Acct-Status-Type = Start",
		"Timestamp = 1537135341",
		"",
		"Mon Sep 17 00:02:22 2018",
		"Acct-Session-Id = \"00201c14283a008f00841b9ed2ed55ca02c7\"",
		"Acct-Status-Type = Stop",
		"Timestamp = 1537135342",
		""}

	metric, err := p.ParseLine(data[0])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[1])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[2])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[3])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[4])
	testutil.RequireMetricEqual(t,
		testutil.MustMetric(
			"radius",
			map[string]string{},
			map[string]interface{}{
				"Timestamp":        int64(1537135341),
				"Acct-Session-Id":  "00201c14283a008f00841b9ed2ed55ca02c7",
				"Acct-Status-Type": "Start",
			},
			time.Unix(1537135341, 0),
		), metric)

	metric, err = p.ParseLine(data[5])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[6])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[7])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[8])
	require.NoError(t, err)
	require.Equal(t, nil, metric)

	metric, err = p.ParseLine(data[9])
	testutil.RequireMetricEqual(t,
		testutil.MustMetric(
			"radius",
			map[string]string{},
			map[string]interface{}{
				"Timestamp":        int64(1537135342),
				"Acct-Session-Id":  "00201c14283a008f00841b9ed2ed55ca02c7",
				"Acct-Status-Type": "Stop",
			},
			time.Unix(1537135342, 0),
		), metric)

}

func BenchmarkParser(b *testing.B) {
	p := RadiusParser{
		MeasurementName: "radius",
		TimestampColumn: "Timestamp",
		TimestampFormat: "Mon Jan 2 15:04:05 2006",
		TimeFunc:        DefaultTime,
	}
	var data = []string{"Mon Sep 17 00:02:21 2018",
		"Acct-Session-Id = \"00201c14283a008f00841b9ed2ed55ca02c7\"",
		"Acct-Status-Type = Start",
		"Timestamp = 1537135341",
		"",
		"Mon Sep 17 00:02:22 2018",
		"Acct-Session-Id = \"00201c14283a008f00841b9ed2ed55ca02c7\"",
		"Acct-Status-Type = Stop",
		"Timestamp = 1537135342",
		""}

	for n := 0; n < b.N; n++ {
		for _, line := range data {
			_, _ = p.ParseLine(line)
		}
	}
}
