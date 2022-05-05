package analyzer

import (
	"fmt"
	"testing"
)

func Test_nTop(t *testing.T) {
	ntr, err := newNTopRecords("test", 3, 0.0, nil, false, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 0; i < 3; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr.register(rowID, score, text, true)
	}

	if ntr.records[0].rowid != 2 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2].rowid != 0 {
		t.Errorf("rowID does not match!")
		return
	}

	ntr, err = newNTopRecords("test", 3, 0.0, nil, false, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 3; i < 5; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr.register(rowID, score, text, true)
	}

	if ntr.records[0].rowid != 4 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2] != nil {
		t.Errorf("Must be nil!")
		return
	}

	ntr, err = newNTopRecords("test", 3, 0.0, nil, false, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 1; i <= 9; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("%03d", i)
		ntr.register(rowID, score, text, true)
	}
	if ntr.records[0].rowid != 9 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2].rowid != 7 {
		t.Errorf("rowID does not match!")
		return
	}

	tran, _ := newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err = newNTopRecords("test", 9, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 1; i <= 9; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		text := fmt.Sprintf("i%02d i%02d i%02d i%02d i%02d",
			i, i+1, i+2, i+3, i+4)
		ntr.register(rowID, score, text, true)
	}
	if ntr.records[0].rowid != 9 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[0].count != 9 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[1] != nil {
		t.Errorf("Must be nil!")
		return
	}

	tran, _ = newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err = newNTopRecords("test", 9, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ntr.register(1, 1.0, "a001 a002 a003 a004 a005", true)
	ntr.register(2, 2.0, "b001 a002 a003 a004 a005", true)
	ntr.register(3, 3.0, "c001 c002 c003 c004 c005", true)
	ntr.register(4, 4.0, "d001 c002 c003 c004 c005", true)
	ntr.register(5, 5.0, "e001 e002 e003 e004 e005", true)
	ntr.register(6, 6.0, "f001 e002 e003 e004 e005", true)

	if ntr.records[0].rowid != 6 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[0].count != 2 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[1].rowid != 4 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[1].count != 2 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[2].rowid != 2 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2].count != 2 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[3] != nil {
		t.Errorf("Must be nil!")
		return
	}

	tran, _ = newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err = newNTopRecords("test", 10, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ntr.register(1, 1.0, "Jun 30 12:08:36 test test: NetScreen device_id=test  [Root]system-alert-00026: IPSec tunnel on interface ethernet0/2 with tunnel ID 0x9a received a packet with a bad SPI. 202.xxx.xxx.154->124.xxx.xxx.186/96, ESP, SPI 0xf5b25b5b, SEQ 0x1. (2021-06-30 12:08:35)", true)
	ntr.register(2, 2.0, "Jun 29 10:58:54 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/128, ESP, SPI 0xf5b259cc, SEQ 0x573. (2021-06-29 10:58:53)", true)
	ntr.register(3, 3.0, "Jun 29 11:12:26 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b259cc, SEQ 0x239e. (2021-06-29 11:12:26)", true)
	ntr.register(4, 4.0, "Jun 29 16:23:20 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/104, ESP, SPI 0xf5b25a1b, SEQ 0x7b73. (2021-06-29 16:23:20)", true)
	ntr.register(5, 5.0, "Jun 29 16:24:12 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b25a1b, SEQ 0x7cfb. (2021-06-29 16:24:11)", true)
	ntr.register(6, 6.0, "Jun 29 16:25:18 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/144, ESP, SPI 0xf5b25a1b, SEQ 0x8142. (2021-06-29 16:25:17)", true)
	ntr.register(7, 7.0, "Jun 29 11:08:50 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of -225 ms. Authentication was None. Update mode was Automatic (2021-06-29 11:08:50)", true)
	ntr.register(8, 8.0, "Jun 30 00:08:53 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of -221 ms. Authentication was None. Update mode was Automatic (2021-06-30 00:08:53)", true)
	ntr.register(9, 9.0, "Jun 29 10:08:50 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of 784 ms. Authentication was None. Update mode was Automatic (2021-06-29 10:08:50)", true)
	ntr.register(10, 10.0, "Jun 29 16:30:03 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b25a1b, SEQ 0x8c4b. (2021-06-29 16:30:03)", true)
	if ntr.records[0].rowid != 10 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[0].count != 6 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[1].rowid != 9 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[1].count != 3 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[2].rowid != 1 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2].count != 1 {
		t.Errorf("count does not match!")
		return
	}

	tran, _ = newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err = newNTopRecords("test", 5, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ntr.register(1, 20.0, "Jun 30 12:08:36 test test: NetScreen device_id=test  [Root]system-alert-00026: IPSec tunnel on interface ethernet0/2 with tunnel ID 0x9a received a packet with a bad SPI. 202.xxx.xxx.154->124.xxx.xxx.186/96, ESP, SPI 0xf5b25b5b, SEQ 0x1. (2021-06-30 12:08:35)", true)
	ntr.register(2, 19.0, "Jun 29 10:58:54 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/128, ESP, SPI 0xf5b259cc, SEQ 0x573. (2021-06-29 10:58:53)", true)
	ntr.register(3, 18.0, "Jun 29 11:12:26 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b259cc, SEQ 0x239e. (2021-06-29 11:12:26)", true)
	ntr.register(4, 17.0, "Jun 29 16:23:20 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/104, ESP, SPI 0xf5b25a1b, SEQ 0x7b73. (2021-06-29 16:23:20)", true)
	ntr.register(5, 16.0, "Jun 29 16:24:12 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b25a1b, SEQ 0x7cfb. (2021-06-29 16:24:11)", true)
	ntr.register(6, 15.0, "Jun 29 16:25:18 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/144, ESP, SPI 0xf5b25a1b, SEQ 0x8142. (2021-06-29 16:25:17)", true)
	ntr.register(7, 14.0, "Jun 29 11:08:50 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of -225 ms. Authentication was None. Update mode was Automatic (2021-06-29 11:08:50)", true)
	ntr.register(8, 13.0, "Jun 30 00:08:53 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of -221 ms. Authentication was None. Update mode was Automatic (2021-06-30 00:08:53)", true)
	ntr.register(9, 12.0, "Jun 29 10:08:50 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type 192.xxx.xxx.1 with an adjustment of 784 ms. Authentication was None. Update mode was Automatic (2021-06-29 10:08:50)", true)
	ntr.register(10, 11.0, "Jun 29 16:30:03 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.xxx.xxx.116 to 124.xxx.xxx.186/96, ESP, SPI 0xf5b25a1b, SEQ 0x8c4b. (2021-06-29 16:30:03)", true)
	ntr.register(11, 10.0, "Jun 29 11:13:05 test test: NetScreen device_id=test  [Root]system-information-00536: IKE 202.pqr.stu.154 Phase 2 msg ID 6b376fe5: Completed negotiations with SPI f5b259d1, tunnel ID 154, and lifetime 3600 seconds/0 KB. (2021-06-29 11:13:04)", true)
	ntr.register(12, 8.0, "test1 test2", true)
	ntr.register(13, 9.0, "Jul  1 00:06:34 test test: NetScreen device_id=test  [Root]system-information-00536: IKE 202.pqr.stu.154 Phase 2 msg ID 2c5008ec: Completed negotiations with SPI f5b25bff, tunnel ID 154, and lifetime 3600 seconds/0 KB. (2021-07-01 00:06:34)", true)

	if ntr.records[0].rowid != 1 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[0].count != 1 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[1].rowid != 9 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[1].count != 3 {
		t.Errorf("count does not match!")
		return
	}
	if ntr.records[2].rowid != 10 {
		t.Errorf("rowID does not match!")
		return
	}
	if ntr.records[2].count != 6 {
		t.Errorf("count does not match!")
		return
	}

	tran, _ = newTrans("", 0, 0, 0, "Jan _2 15:04:05", 1, 0)
	ntr, err = newNTopRecords("test", 10, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ntr.register(1, 20.0, "Jun 30 12:08:36 test test: NetScreen device_id=test  [Root]system-alert-00026: IPSec tunnel on interface ethernet0/2 with tunnel ID 0x9a received a packet with a bad SPI. 202.pqr.stu.154->124.gh.ijk.186/96, ESP, SPI 0xf5b25b5b, SEQ 0x1. (2021-06-30 12:08:35)", true)
	ntr.register(2, 19.0, "Jun 29 10:58:54 test test: NetScreen device_id=test  [Root]system-critical-00042: Replay packet detected on IPSec tunnel on ethernet0/2 with tunnel ID 0x67! From 121.lm.no.116 to 124.gh.ijk.186/128, ESP, SPI 0xf5b259cc, SEQ 0x573. (2021-06-29 10:58:53)", true)
	ntr.register(3, 18.0, "Jun 29 11:08:50 test test: NetScreen device_id=test  [Root]system-notification-00531: The system clock was updated from primary NTP server type abc.def.88.1 with an adjustment of -225 ms. Authentication was None. Update mode was Automatic (2021-06-29 11:08:50)", true)
	ntr.register(4, 17.0, "Jun 29 10:38:20 test test: NetScreen device_id=test  [Root]system-information-00536: Rejected an IKE packet on ethernet0/2 from 216.abc.def.110:53271 to 124.gh.ijk.186:500 with cookies 3e35c70729dfedef and 0000000000000000 because an initial Phase 1 packet arrived from an unrecognized peer gateway. (2021-06-29 10:38:20)", true)
	ntr.register(5, 16.0, "Jun 30 21:48:16 test test: NetScreen device_id=test  [Root]system-information-00767: Save configuration to IP address abc.def.32.133 under filename test-20210630-2148.cfg by administrator by admin brastel_admins. (2021-06-30 21:48:15)", true)
	ntr.register(6, 15.0, "Jun 30 00:10:45 test test: NetScreen device_id=test  [Root]system-information-00536: IKE 202.pqr.stu.154 Phase 2 msg ID 7dbb4944: Completed negotiations with SPI f5b25aa4, tunnel ID 154, and lifetime 3600 seconds/0 KB. (2021-06-30 00:10:45)", true)
	ntr.register(7, 14.0, "Jun 29 11:33:38 test test: NetScreen device_id=test  [Root]system-notification-00257(traffic): start_time='2021-06-29 11:33:35' duration=3 policy_id=920 service=http proto=6 src zone=Trust dst zone=Untrust action=Permit sent=899 rcvd=3582 src=abc.def.199.106 dst=gh.ijk.156.80 src_port=63016 dst_port=80 src-xlated ip=202.214.113.11 port=4152 dst-xlated ip=gh.ijk.156.80 port=80 session_i", true)
	ntr.register(8, 13.0, "Jun 30 10:23:44 test test: NetScreen device_id=test  [Root]system-information-00536: VPN monitoring for VPN IKE52 Manila via IIJ - PLDT has deactivated the SA with ID 0x0000008c. (2021-06-30 10:23:43)", true)
	ntr.register(9, 12.0, "Jun 29 11:13:05 test test: NetScreen device_id=test  [Root]system-information-00536: IKE 202.pqr.stu.154 Phase 2 msg ID 6b376fe5: Completed negotiations with SPI f5b259d1, tunnel ID 154, and lifetime 3600 seconds/0 KB. (2021-06-29 11:13:04)", true)
	ntr.register(10, 11.0, "Jul  1 00:06:34 test test: NetScreen device_id=test  [Root]system-information-00536: IKE 202.pqr.stu.154 Phase 2 msg ID 2c5008ec: Completed negotiations with SPI f5b25bff, tunnel ID 154, and lifetime 3600 seconds/0 KB. (2021-07-01 00:06:34)", true)

	if ntr.records[0].rowid != 1 {
		t.Errorf("rowID does not match!")
		return
	}

	r := ntr.getRecords()
	if r[0].rowid != 1 {
		t.Errorf("rowID does not match!")
		return
	}

	if r[7].rowid != 10 {
		t.Errorf("rowID does not match!")
		return
	}

	if r[7].count != 3 {
		t.Errorf("count does not match!")
		return
	}
}

func Test_nTop2(t *testing.T) {
	tran, _ := newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err := newNTopRecords("test", 10, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 1; i <= 80; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		j := 0
		if i%3 == 0 {
			j = 1
		}
		text := fmt.Sprintf("i%02d i%02d i%02d i%02d i%02d",
			i, i+1+j, i+2+j, i+3+j, i+4+j)

		ntr.register(rowID, score, text, true)
	}
	r := ntr.getRecords()
	if len(r) != 10 {
		t.Errorf("length does not match!")
		return
	}
	if r[0].rowid != 80 {
		t.Errorf("rowID does not match!")
		return
	}

	tran, _ = newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err = newNTopRecords("test", 10, 0.0, tran, true, "", 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 1; i <= 3; i++ {
		rowID := int64(i)
		score := float64(i) * 2.0
		j := 0
		if i%3 == 0 {
			j = 1
		}
		text := fmt.Sprintf("i%02d i%02d i%02d i%02d i%02d",
			i, i+1+j, i+2+j, i+3+j, i+4+j)

		ntr.register(rowID, score, text, true)
	}
	r = ntr.getRecords()
	if len(r) != 2 {
		t.Errorf("length does not match!")
		return
	}
	if r[1] == nil {
		t.Errorf("Must not be nil!")
		return
	}
}

func Test_nTop3(t *testing.T) {
	err := removeTestDir("Test_nTop3")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	dataDir, err := initTestDir("Test_nTopDiff")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tran, _ := newTrans("", 0, 0, 0, "Jan _2 15:04:05", 1, 0)
	ntr, err := newNTopRecords("test", 10, 0.0, tran, true, dataDir, 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	ntr.register(1, 1.0, "Oct 11 01:18:14 te101 te102 te103 te104 te105", true)
	ntr.register(2, 1.1, "Oct 11 02:19:14 ty101 ty102 ty103 te104 te105", true)
	ntr.register(3, 1.2, "Oct 11 02:20:14 ty101 ty102 ty103 te104 te105", true)
	ntr.register(4, 1.0, "Oct 11 03:21:14 wk101 wk102 wk103 wk104 wk105", true)

	rs := ntr.getRecords()
	if rs[1].rowid != 4 {
		t.Errorf("row id incorrect")
		return
	}

	ntr.register(5, 1.1, "Oct 11 03:22:14 wk101 wk102 wk103 wk104 wk105", true)
	ntr.register(6, 1.2, "Oct 11 03:23:14 wk101 wk102 wk103 wk104 wk105", true)
	ntr.register(7, 1.3, "Oct 11 03:24:14 wk101 wk102 wk103 wk104 wk105", true)

	rs = ntr.getRecords()
	if rs[0].rowid != 7 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[1].rowid != 3 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[2].rowid != 1 {
		t.Errorf("row id incorrect")
		return
	}

	if err := ntr.save(); err != nil {
		t.Errorf("%v", err)
		return
	}

	ntr, err = newNTopRecords("test", 10, 0.0, tran, true, dataDir, 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := ntr.load(12, 10, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	rs = ntr.getRecords()
	if rs[0].rowid != 7 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[1].rowid != 3 {
		t.Errorf("row id incorrect")
		return
	}

	ntr.register(12, 1.0, "Oct 12 01:18:14 te101 te102 te103 te104 te105", true)
	rs = ntr.getRecords()
	if rs[0].rowid != 7 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[1].rowid != 3 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[2].rowid != 12 {
		t.Errorf("row id incorrect")
		return
	}

	ntr.register(13, 1.3, "Oct 12 02:19:14 ty101 ty102 ty103 te104 te105", true)
	ntr.register(14, 1.4, "Oct 12 02:20:14 ty101 ty102 ty103 te104 te105", true)
	ntr.register(15, 1.0, "Oct 12 03:21:14 wk101 wk102 wk103 wk104 wk105", true)

	rs = ntr.getRecords()
	if rs[0].rowid != 14 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[1].rowid != 15 {
		t.Errorf("row id incorrect")
		return
	}
	if rs[1].maxScore != 1.3 {
		t.Errorf("max score incorrect")
		return
	}
	if rs[2].rowid != 12 {
		t.Errorf("row id incorrect")
		return
	}

}

func Test_nTop4(t *testing.T) {
	dataDir, err := initTestDir("Test_nTop4")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tran, _ := newTrans("", 0, 0, 0, "", 1, 0)
	ntr, err := newNTopRecords("test", 10, 0.0, tran, true, dataDir, 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	for i := 0; i < 200; i++ {
		rowID := int64(i)
		score := float64(i)
		text := fmt.Sprintf("i%03d", i)
		ntr.register(rowID, score, text, true)
	}
	if ntr.records[0].rowid != 199 {
		t.Errorf("row id incorrect")
		return
	}
	if ntr.records[99].rowid != 100 {
		t.Errorf("row id incorrect")
		return
	}
	recs := ntr.getRecords()
	if recs[9].rowid != 190 {
		t.Errorf("row id incorrect")
		return
	}

	if err := ntr.save(); err != nil {
		t.Errorf("%v", err)
		return
	}

	ntr, err = newNTopRecords("test", 10, 0.0, tran, true, dataDir, 0)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if err := ntr.load(200, 1000, true); err != nil {
		t.Errorf("%v", err)
		return
	}

	ntr.register(201, 99.0, "i201", true)
	for _, rec := range ntr.records {
		if rec.rowid == 201 {
			t.Errorf("Must not include this rowid")
			return
		}
	}

	ntr.register(202, 189.0, "i202", true)
	if ntr.records[10].rowid != 202 {
		t.Errorf("row id incorrect")
		return
	}
}
