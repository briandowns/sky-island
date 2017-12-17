package jail

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
// <monitoring jail>/etc/rc.conf file
func (j *jailService) updateMonitoringRcConf() error {
	etcConf, err := os.OpenFile(j.conf.Jails.BaseJailDir+"/monitoring/etc/rc.conf", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer etcConf.Close()
	etcConf.Write([]byte("\ninfluxd_enable=\"YES\"\n"))
	etcConf.Write([]byte("grafana_enable=\"YES\"\n"))
	etcConf.Write([]byte("telegraf_enable=\"YES\"\n"))
	return nil
}

// disableSendmail sets all necessary parameters in the /etc/rc.conf file
// to make sure that sendmail(8) isn't started
func (j *jailService) disableMonitoringSendmail() error {
	etcConf, err := os.OpenFile(j.conf.Jails.BaseJailDir+"/monitoring/etc/rc.conf", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer etcConf.Close()
	etcConf.Write([]byte("\nsendmail_enable=\"NO\"\n"))
	etcConf.Write([]byte("sendmail_submit_enable=\"NO\"\n"))
	etcConf.Write([]byte("sendmail_outbound_enable=\"NO\"\n"))
	etcConf.Write([]byte("sendmail_msp_queue_enable=\"NO\"\n"))
	return nil
}

// writeConfig takes the given data and writes it to the file
func writeConfig(confFile string, data []byte) error {
	ic, err := os.Create(confFile)
	if err != nil {
		return err
	}
	defer ic.Close()
	ic.Write(data)
	return nil
}

var monitoringConfigFiles = []string{"influxd.conf", "telegraf.conf", "grafana.conf"}

// buildMonitoringJail creates, configures, and starts
// the monitoring jail
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
	if err := j.disableMonitoringSendmail(); err != nil {
		return err
	}
	j.logger.Log("msg", "Starting monitoring jail")
	res0, err := j.wrapper.CombinedOutput("jail", "-c", "monitoring")
	if err != nil {
		return errors.New(string(res0))
	}
	j.logger.Log("msg", "Installing pkg in monitoring jail")
	res1, err := j.wrapper.CombinedOutput("pkg", "-j", "monitoring", "install", "-y", "pkg")
	if err != nil {
		return errors.New(string(res1))
	}
	j.logger.Log("msg", "Installing influx, etc... in monitoring jail")
	res2, err := j.wrapper.CombinedOutput("pkg", "-j", "monitoring", "install", "-y", "influxdb", "telegraf", "grafana4")
	if err != nil {
		return errors.New(string(res2))
	}
	// most of the calls below can be done concurrently
	j.logger.Log("msg", "Updating /etc/rc.conf in monitoring jail")
	if err := j.updateMonitoringRcConf(); err != nil {
		return err
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
	cmd := exec.Command("jail", "-rc", "monitoring")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
