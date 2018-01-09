var statusSource = document.querySelector("#status-template").innerHTML;
var statusTemplate = Handlebars.compile(statusSource);

var jailsSource = document.querySelector("#jails-template").innerHTML;
var jailsTemplate = Handlebars.compile(jailsSource);

var statusModel = {
    systemStatus: 'UNKNOWN',
    systemStatusStyle: 'alert-secondary',
    uptime: 'UNKNOWN',
    count: 'UNKNOWN',
    totalResponseTime: 'UNKNOWN',
    averageResponseTime: 'UNKNOWN',
    statusCodes: 'UNKNOWN',
    asOfTime: 'UNKNOWN'
};

var jailsModel = {
    jails: []
};

function updateStatusTemplate() {
    var html = statusTemplate(statusModel);
    document.querySelector('#status-insertion-point').innerHTML = html;
}

updateStatusTemplate();

function updateJailsTemplates() {
    var html = jailsTemplate(jailsModel);
    document.querySelector('#jails-insertion-point').innerHTML = html;
}

updateJailsTemplates();

function pollHealthCheck() {
    $.get('./healthcheck', function (data, status) {
        if (status == 'success') {
            statusModel.systemStatus = 'OK';
            statusModel.systemStatusStyle = 'alert-success';
        } else {
            console.error(data);
            console.error(status);

            statusModel.systemStatus = 'ERROR';
            statusModel.systemStatusStyle = 'alert-danger';
        }

        updateStatusTemplate();
        setTimeout(pollHealthCheck, 60 * 1000);
    });
}

pollHealthCheck();


function pollApiStats() {
    //TODO do something
    $.get({
        url: './api/v1/admin/api-stats',
        headers: {
            'X-Sky-Island-Token': 'asdfasdfasdfasdf'
        }
    }, function (data, status) {
        if (status == 'success') {
            statusModel.uptime = data.uptime;
            statusModel.count = data.total_count;
            statusModel.totalResponseTime = data.total_response_time;
            statusModel.averageResponseTime = data.average_response_time;
            statusModel.statusCodes = Object
                .entries(data.total_status_code_count)
                .map(function (entry) {
                    return {
                        code: entry[0],
                        count: entry[1]
                    };
                });
            statusModel.asOfTime = data.time;

            updateStatusTemplate();
        } else {
            console.error(data);
            console.error(status);
        }
    });

    setTimeout(pollApiStats, 60 * 1000);
}

pollApiStats();

function pollJails() {
    $.get({
        url: './api/v1/admin/jails',
        headers: {
            'X-Sky-Island-Token': 'asdfasdfasdfasdf'
        }
    }, function (data, status) {
        if (!data.jails) {
            return;
        }

        jailsModel = data.jails.map(function (jail) {
            return {
                host: jail.host,
                ip4: jail.ip4,
                ip6: jail.ip6,
                jid: jail.jid,
                name: jail.name,
                os: jail.OSRelease,
                path: jail.path
            };
        });

        updateJailsTemplates();
    });

    setTimeout(pollJails, 60 * 1000);
}

pollJails();
