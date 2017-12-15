package jail

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/utils"
	"gopkg.in/alexcesaro/statsd.v2"
)

var testConf = &config.Config{
	Network: &config.Network{
		IP4: &config.IP4{
			Interface: "em0",
			StartAddr: "192.168.0.20",
			Range:     220,
		},
	},
	Jails: &config.Jails{
		BaseJailDir:            "",
		CacheDefaultExpiration: "24h",
		CachePurgeAfter:        "48h",
	},
}

// TestNewJailService
func TestNewJailService(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestDownloadBaseSystem
func TestDownloadBaseSystem(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}

}

// TestExtractBasePkgs
func TestExtractBasePkgs(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestUpdateBaseJail
func TestUpdateBaseJail(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestSetBaseJailConf
func TestSetBaseJailConf(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestConfigureJailHostname
func TestConfigureJailHostname(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestDownloadGo
func TestDownloadGo(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestInitializeSystem
func TestInitializeSystem(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestCreateJail
func TestCreateJail(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

// TestRemoveJail
func TestRemoveJail(t *testing.T) {
	jailSvc := NewJailService(testConf, &logrus.Logger{}, &statsd.Client{}, utils.NoOpWrapper{})
	if jailSvc == nil {
		t.Error("expected not nil jail service")
	}
}

var testJLSData = `devfs_ruleset=0 enforce_statfs=2 host=new ip4=disable ip6=disable jid=67 name=f7bb0cf4-caf2-11e7-ac45-0800279d94cc osreldate=1101001 osrelease=11.1-RELEASE path=/zroot/jails/f7bb0cf4-caf2-11e7-ac45-0800279d94cc persist securelevel=-1 sysvmsg=disable sysvsem=disable sysvshm=disable allow.nochflags allow.nomount allow.mount.nodevfs allow.mount.nofdescfs allow.mount.nolinprocfs allow.mount.nolinsysfs allow.mount.nonullfs allow.mount.noprocfs allow.mount.notmpfs allow.mount.nozfs allow.noquotas allow.noraw_sockets allow.set_hostname allow.nosocket_af allow.nosysvipc children.max=0 host.domainname="" host.hostid=0 host.hostname=f7bb0cf4-caf2-11e7-ac45-0800279d94cc host.hostuuid=00000000-0000-0000-0000-000000000000
devfs_ruleset=0 enforce_statfs=2 host=new ip4=disable ip6=disable jid=67 name=f7bb0cf4-caf2-11e7-ac45-0800279d94cc osreldate=1101001 osrelease=11.1-RELEASE path=/zroot/jails/f7bb0cf4-caf2-11e7-ac45-0800279d94cc persist securelevel=-1 sysvmsg=disable sysvsem=disable sysvshm=disable allow.nochflags allow.nomount allow.mount.nodevfs allow.mount.nofdescfs allow.mount.nolinprocfs allow.mount.nolinsysfs allow.mount.nonullfs allow.mount.noprocfs allow.mount.notmpfs allow.mount.nozfs allow.noquotas allow.noraw_sockets allow.set_hostname allow.nosocket_af allow.nosysvipc children.max=0 host.domainname="" host.hostid=0 host.hostname=f7bb0cf4-caf2-11e7-ac45-0800279d94cc host.hostuuid=00000000-0000-0000-0000-000000000000`

// TestJLSRun_Success
func TestJLSRun_Success(t *testing.T) {
	_, err := JLSRun(utils.NoOpWrapper{})
	if err != nil {
		t.Log(err)
		t.Error(err)
	}
}

//1330301281001/1330301281001_4651606561001_4651453390001.zip

/*
{
  "health_check_address": ":8107",
  "statsd_address": "127.0.0.1:8125",
  "statsd_reporting_interval_seconds": 5,
  "fabric": {
    "base_url": "https://management.us-east-1.prod.boltdns.net",
    "media_url": "https://media.us-east-1.prod.boltdns.net",
    "jaws_url": "https://jaws.us-east-1.prod.boltdns.net",
    "auth_token": "1e98db83-22b2-4bf9-b637-abe57c62610e"
  },
  "castlabs": {
    "user": "brightcove::packaging",
    "pass": "RW2pCmYaNv7AdkkR",
    "service": "ingest",
    "environment": "prod"
  },
  "dynamo": {
    "dynamo_client": {
      "region": "us-east-1"
    },
    "dynamodb_table_title": {
      "table_name": "prod-state-management-TitleTable-BPXB45352HLS"
    }
  },
  "swf": {
    "poller_thread_count": 10,
    "media_processor_poller_thread_count": 10,
    "domain": "bolt-jaws-prod-a",
    "task_list": ["Go-Delta", "Go-Echo", "Go-Foxtrot"],
    "region": "us-east-1"
  },
  "s3": {
    "region": "us-east-1",
    "bucket": "not-needed-but-required-by-config-schema"
  },
  "s3_buckets": {
    "buckets": {
      "origin": "bolt-media.prod.%s",
      "archive": "bolt-mezzanine.prod.%s"
    }
  },
  "roebuck": {
    "base_url": "http://roebuck-rest.kar.brightcove.com",
    "file_store_to_bucket": {
      "FIJI": "brightcove-us-archive-1-us-east-1",
      "NAS1": "brightcove-us-archive-2-us-east-1",
      "NAS2": "brightcove-us-archive-3-us-east-1",
      "NAS3": "brightcove-us-archive-4-us-east-1",
      "NAS4": "brightcove-us-archive-2-us-east-1",
      "NAS7": "brightcove-us-archive-4-us-east-1",
      "FIJIFALKLAND2": "brightcove-us-archive-4-us-east-1",
      "S3DM1": "brightcove-digitalmasters-us-east-1",
      "S3_US_ARCHIVE_1_US_EAST_1": "brightcove-us-archive-1-us-east-1",
      "S3_US_ARCHIVE_2_US_EAST_1": "brightcove-us-archive-2-us-east-1",
      "S3_US_ARCHIVE_3_US_EAST_1": "brightcove-us-archive-3-us-east-1",
      "S3_US_ARCHIVE_4_US_EAST_1": "brightcove-us-archive-4-us-east-1",
      "MANNHEIMS3": "brightcove-mannheim-storage-us-east-1",
      "LIVEVODS3": "brightcove-livetovod-perm-us-east-1"
    }
  },
  "bolt": {
    "auth_token": "87ea83f3-ab37-4a7d-98af-a4df20803035"
   },
  "banshi": {
    "base_url": "http://internal-banshi-api-lb-1140481306.us-east-1.elb.amazonaws.com",
    "token": "853dbad0-f232-11e5-ba58-2188b824b7cd"
  },
  "gc_watcher_conf": {
    "schedule": "* * * * *"
  },
  "swf_watcher_conf": {
    "schedule": "* * * * *"
  },
  "working_dir": "/data"
}
*/
