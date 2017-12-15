package jail

import (
	"errors"
	"fmt"
	"os"
)

const monitoringJailConf = `
monitoring {
	path = "%s/${name}";
	devfs_ruleset = 4;
	host.hostname = ${name};
	ip4.addr = "%s|%s";
	ip4.addr += "lo0|127.0.0.2/8";
	exec.start = "/bin/sh /etc/rc";
	exec.stop = "/bin/sh /etc/rc.shutdown";
}
`

const influxConf = `
[meta]
	dir = "/var/db/influxdb/meta"
	logging-enabled = true

[data]
	dir = "/var/db/influxdb/data"
	wal-dir = "/var/db/influxdb/wal"
	query-log-enabled = true

[monitor]
	store-enabled = true
	store-database = "telegraf"

[http]
	enabled = true
	bind-address = ":8086"
`

const grafanaConf = `
[paths]
	data = /var/db/grafana/
	logs = /var/log/grafana/
	plugins = /var/db/grafana/plugins

[server]
	static_root_path = public

[database]
	type = sqlite3
	path = grafana.db

[analytics]
	check_for_updates = true

[dashboards.json]
	path = /var/db/grafana/dashboards
`

const telegrafConf = `
[[inputs.statsd]]
	service_address = ":8125"
	delete_gauges = true
	delete_counters = true
	delete_sets = true
	delete_timings = true
	percentiles = [90]
	metric_separator = "_"
	parse_data_dog_tags = false
	allowed_pending_messages = 10000
	percentile_limit = 1000
`

// updateMonitoringRcConf updates the monitoring jail's
// /etc/rc.conf file
func updateMonitoringRcConf() error {
	etcConf, err := os.OpenFile(rcConf, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer etcConf.Close()
	etcConf.Write([]byte("influxd_enable=\"YES\"\n"))
	etcConf.Write([]byte("grafana_enable=\"YES\"\n"))
	etcConf.Write([]byte("telegraf_enable=\"YES\"\n"))
	return nil
}

// writeConfig
func writeConfig(confFile string, data []byte) error {
	ic, err := os.OpenFile(confFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer ic.Close()
	ic.Write(data)
	return nil
}

// buildMonitoringJail
func (j *jailService) buildMonitoringJail() error {
	if err := j.CreateJail("monitoring", false); err != nil {
		return err
	}
	f, err := os.OpenFile("/etc/jail.conf", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	jc := fmt.Sprintf(monitoringJailConf, j.conf.Jails.BaseJailDir, j.conf.Network.IP4.Interface, j.conf.Jails.MonitoringAddr)
	if _, err = f.WriteString(jc); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := disableSendmail(); err != nil {
		return err
	}
	res0, err := j.wrapper.CombinedOutput("jail", "-c", "monitoring")
	if err != nil {
		return errors.New(string(res0))
	}
	res1, err := j.wrapper.CombinedOutput("pkg", "-j", "monitoring", "install", "-y", "pkg")
	if err != nil {
		return errors.New(string(res1))
	}
	res2, err := j.wrapper.CombinedOutput("pkg", "-j", "monitoring", "install", "-y", "influxdb", "telegraf", "grafana4")
	if err != nil {
		return errors.New(string(res2))
	}
	if err := updateMonitoringRcConf(); err != nil {
		return errors.New(string(res2))
	}
	ic := j.conf.Jails.BaseJailDir + "/monitoring/usr/local/etc/influxd.conf"
	if err := writeConfig(ic, []byte(influxConf)); err != nil {
		return err
	}
	tc := j.conf.Jails.BaseJailDir + "/monitoring/usr/local/etc/telegraf.conf"
	if err := writeConfig(tc, []byte(telegrafConf)); err != nil {
		return err
	}
	gc := j.conf.Jails.BaseJailDir + "/monitoring/usr/local/etc/grafana.conf"
	if err := writeConfig(gc, []byte(grafanaConf)); err != nil {
		return err
	}
	res3, err := j.wrapper.CombinedOutput("jail", "-rc", "monitoring")
	if err != nil {
		return errors.New(string(res3))
	}
	return nil
}
