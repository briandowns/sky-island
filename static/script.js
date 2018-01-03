function pollHealthCheck() {
    $.get('http://demo.skyisland.io:3280/healthcheck', function (data, status) {
        if (status == 'success') {
            $('#healthcheck').removeClass('alert-failure alert-secondary').addClass('alert-success');
            $('#healthcheck-text').text('OK');
        } else {
            console.error(data);
            console.error(status);

            $('#healthcheck').removeClass('alert-success alert-secondary').addClass('alert-danger');
            $('#healthcheck-text').text('ERROR');
        }

        setTimeout(pollHealthCheck, 60 * 1000);
    });
}

pollHealthCheck();


function pollApiStats() {
    //TODO do something
    $.get({
        url: 'http://demo.skyisland.io:3280/api/v1/admin/api-stats',
        headers: {
            'X-Sky-Island-Token': 'asdfasdfasdfasdf'
        }
    }, function (data, status) {
        if (status == 'success') {
            $('#uptime').text(data.uptime);
            $('#count').text(data.total_count);
            $('#total_response_time').text(data.total_response_time);
            $('#average_response_time').text(data.average_response_time);

            var counts = data.total_status_code_count;
            var html = '';
            for (var key in counts) {
                html += key + ' ' + counts[key] + '<br />';
            }

            $('#status_codes').html(html);

            $('#time').text(data.time);
        } else {
            console.error(data);
            console.error(status);
        }
    });

    setTimeout(pollApiStats, 60 * 1000);
}

pollApiStats();

function getJailBlock(name, stats) {
    var statsHtml = [];

    for (var key in stats) {
        statsHtml.push(
            [   
                '<tr>',
                    '<td>'+ key +'</td>',
                    '<td>'+ stats[key] +'</td>',
                '</tr>'
            ].join('')
        );
    }

    var jailBlock = [
        '<div class="panel panel-default">',
            '<div class="panel-body">',
                '<h4>b6ada3ce-ede7-11e7-b0fe-12e74a189a08</h4>',
                '<table class="table">',
                statsHtml,
                '</table>',
            '</div>',
        '</div>'
    ];

    return jailBlock.join('');
}
function pollJails() {
    $.get({
        url: 'http://demo.skyisland.io:3280/api/v1/admin/jails',
        headers: {
            'X-Sky-Island-Token': 'asdfasdfasdfasdf'
        }
    }, function(data, status) {
        var html = data.jails.map(function(jail) {
            return getJailBlock(jail.name, jail);
        }).join('');

        console.log('appending!');

        $('#jails').append(html);
    });

    setTimeout(pollJails, 60 * 1000);

    /*
    {
        "jails": [
            {
                "host": "new",
                "ip4": "disable",
                "ip6": "disable",
                "jid": 188,
                "name": "b6ada3ce-ede7-11e7-b0fe-12e74a189a08",
                "OSRelease": "11.1-RELEASE-p4",
                "path": "/zroot/jails/b6ada3ce-ede7-11e7-b0fe-12e74a189a08",
                "hostname": "b6ada3ce-ede7-11e7-b0fe-12e74a189a08"
            }
        ]
    }
    */
}

pollJails();