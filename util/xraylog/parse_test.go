package xraylog_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v2/util/xraylog"
)

func TestRecordParsers_ManyVariants(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantFrom    string
		wantTo      string
		wantEmail   string
		wantAddrOk  bool
		wantEmailOk bool
	}{
		// ---------- BASIC ----------
		{
			name:       "tcp ip to ip",
			line:       "from 127.0.0.1:1234 accepted tcp:127.0.0.1:443",
			wantFrom:   "127.0.0.1:1234",
			wantTo:     "tcp:127.0.0.1:443",
			wantAddrOk: true,
		},
		{
			name:       "udp ip to ip",
			line:       "from 10.0.0.1:5555 accepted udp:10.0.0.2:53",
			wantFrom:   "10.0.0.1:5555",
			wantTo:     "udp:10.0.0.2:53",
			wantAddrOk: true,
		},

		// ---------- SLASHES ----------
		{
			name:       "tcp with double slashes",
			line:       "from 1.1.1.1:1111 accepted tcp://example.com:443",
			wantFrom:   "1.1.1.1:1111",
			wantTo:     "tcp://example.com:443",
			wantAddrOk: true,
		},
		{
			name:       "udp with slashes and path",
			line:       "from 2.2.2.2:2222 accepted udp://host.local:53/dns",
			wantFrom:   "2.2.2.2:2222",
			wantTo:     "udp://host.local:53/dns",
			wantAddrOk: true,
		},

		// ---------- DOMAINS ----------
		{
			name:       "domain with port",
			line:       "from localhost:6000 accepted tcp:api.service.local:8443",
			wantFrom:   "localhost:6000",
			wantTo:     "tcp:api.service.local:8443",
			wantAddrOk: true,
		},

		// ---------- EMAIL SIMPLE ----------
		{
			name:        "simple email",
			line:        "from 3.3.3.3:3333 accepted tcp:google.com:443 email: test",
			wantFrom:    "3.3.3.3:3333",
			wantTo:      "tcp:google.com:443",
			wantEmail:   "test",
			wantAddrOk:  true,
			wantEmailOk: true,
		},

		// ---------- EMAIL COMPLEX ----------
		{
			name:        "email with symbols",
			line:        "from 4.4.4.4:4444 accepted tcp:example.com:443 email: user.name+dev_123@test-domain.io",
			wantFrom:    "4.4.4.4:4444",
			wantTo:      "tcp:example.com:443",
			wantEmail:   "user.name+dev_123@test-domain.io",
			wantAddrOk:  true,
			wantEmailOk: true,
		},
		{
			name:        "email at end",
			line:        "from 5.5.5.5:5555 accepted tcp:example.com:443 email: a_b-c.d+e@x.y",
			wantFrom:    "5.5.5.5:5555",
			wantTo:      "tcp:example.com:443",
			wantEmail:   "a_b-c.d+e@x.y",
			wantAddrOk:  true,
			wantEmailOk: true,
		},

		// ---------- NO EMAIL ----------
		{
			name:        "no email",
			line:        "from 6.6.6.6:6666 accepted tcp:example.com:443",
			wantFrom:    "6.6.6.6:6666",
			wantTo:      "tcp:example.com:443",
			wantAddrOk:  true,
			wantEmailOk: false,
		},

		// ---------- BROKEN ----------
		{
			name:        "broken accepted",
			line:        "from 7.7.7.7:7777 tcp:example.com:443",
			wantFrom:    "7.7.7.7:7777",
			wantAddrOk:  false,
			wantEmailOk: false,
		},
		{
			name:        "broken from",
			line:        "accepted tcp:example.com:443 email: bad",
			wantTo:      "tcp:example.com:443",
			wantAddrOk:  false,
			wantEmailOk: true,
			wantEmail:   "bad",
		},

		// ---------- GARBAGE ----------
		{
			name:        "garbage line",
			line:        "abracadabra",
			wantAddrOk:  false,
			wantEmailOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := []byte(tt.line)

			from, to, addrOk := xraylog.RecordAddrs(rec)
			if addrOk != tt.wantAddrOk {
				t.Fatalf("addrOk = %v, want %v", addrOk, tt.wantAddrOk)
			}
			if from != tt.wantFrom {
				t.Errorf("from = %q, want %q", from, tt.wantFrom)
			}
			if to != tt.wantTo {
				t.Errorf("to = %q, want %q", to, tt.wantTo)
			}

			email, emailOk := xraylog.RecordEmail(rec)
			if emailOk != tt.wantEmailOk {
				t.Fatalf("emailOk = %v, want %v", emailOk, tt.wantEmailOk)
			}
			if email != tt.wantEmail {
				t.Errorf("email = %q, want %q", email, tt.wantEmail)
			}
		})
	}
}

func TestRecordsExample(t *testing.T) {
	logs := `2026/01/21 18:51:04.003185 from 127.0.0.1:35256 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:08.035307 from 127.0.0.1:35270 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:14.002851 from 127.0.0.1:35340 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:24.003191 from 127.0.0.1:37870 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:25.849264 from 10.192.0.100:3022 accepted tcp:www.google.com:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:51:27.046418 from 10.192.0.100:3029 accepted tcp:www.google.com:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:51:33.816162 from 10.192.0.100:3050 accepted tcp:www.google.com:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:51:34.001476 from 127.0.0.1:36910 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:44.002439 from 127.0.0.1:43848 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:51:54.003371 from 127.0.0.1:59142 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:04.002864 from 127.0.0.1:44570 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:13.215453 from 10.192.0.100:3108 accepted tcp:www.google.com:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:52:13.841165 from 10.192.0.100:3110 accepted tcp:cachefly.cachefly.net:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:52:14.003567 from 127.0.0.1:38482 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:24.003055 from 127.0.0.1:39834 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:34.002930 from 127.0.0.1:35174 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:41.649467 from 10.192.0.100:3130 accepted tcp:www.google.com:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:52:42.261083 from 10.192.0.100:3132 accepted tcp:cachefly.cachefly.net:443 [inbound-1080 >> direct] email: ftzqt5e2
2026/01/21 18:52:44.001208 from 127.0.0.1:45334 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:52:54.003235 from 127.0.0.1:53764 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:04.003383 from 127.0.0.1:40466 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:14.002281 from 127.0.0.1:56164 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:24.002534 from 127.0.0.1:58872 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:34.002281 from 127.0.0.1:54620 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:44.002315 from 127.0.0.1:39382 accepted tcp:127.0.0.1:62789 [api -> api]
2026/01/21 18:53:54.003315 from 127.0.0.1:34552 accepted tcp:127.0.0.1:62789 [api -> api]
`

	scanner := bufio.NewScanner(strings.NewReader(logs))

	var (
		totalLines  int
		apiSkipped  int
		parsed      int
		withEmail   int
		emailValues = map[string]int{}
	)

	var filterMatches int = 0

	for scanner.Scan() {
		totalLines++
		record := scanner.Bytes()

		if xraylog.RecordIsApiCallOrEmpty(record) {
			apiSkipped++
			continue
		}

		if xraylog.FoundInRecordByFilterString(record, "www.google.com") {
			filterMatches++
		}

		if !xraylog.FoundInRecordByFilterString(record, "") {
			t.Fatalf("record must not be filtered: %s", record)
		}

		if _, ok := xraylog.RecordTimestamp(record); !ok {
			t.Fatalf("timestamp not parsed: %s", record)
		}

		inb, outb, ok := xraylog.RecordRoute(record)
		if !ok {
			t.Fatalf("route not parsed: %s", record)
		}
		if inb == "" || outb == "" {
			t.Fatalf("empty route: %s", record)
		}

		from, to, ok := xraylog.RecordAddrs(record)
		if !ok {
			t.Fatalf("addrs not parsed: %s", record)
		}
		if from == "" || to == "" {
			t.Fatalf("empty addr: %s", record)
		}

		if email, ok := xraylog.RecordEmail(record); ok {
			withEmail++
			emailValues[email]++
		}

		parsed++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	// ---------- ASSERTS ----------

	if filterMatches != 5 {
		t.Fatalf("filter must match 5 lines, but matched = %d", filterMatches)
	}

	if totalLines != 26 {
		t.Fatalf("totalLines = %d, want 26", totalLines)
	}

	if apiSkipped == 18 {
		t.Fatalf("apiSkipped must be > 0 and < total")
	}

	if parsed == 0 {
		t.Fatal("no records parsed")
	}

	if withEmail != 7 {
		t.Fatalf("withEmail = %d, want 7", withEmail)
	}

	if len(emailValues) != 1 {
		t.Fatalf("emails = %+v, want exactly one unique", emailValues)
	}

	if emailValues["ftzqt5e2"] != 7 {
		t.Fatalf("email ftzqt5e2 count = %d, want 7", emailValues["ftzqt5e2"])
	}
}
