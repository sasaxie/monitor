var infoUrl = serverHost + "/v1/monitor/info/tag/";
var wsInfoUrl = wsServerHost + "/v1/wsmonitor/tag";
var settingsUrl = serverHost + "/v1/monitor/settings/";
var runTimeUrl = serverHost + "/v1/monitor/program-info/";

// 页面加载后执行
$(document).ready(function () {
    initTag();

    initRunTime();

    var table = $('#showdatatable').DataTable({
        destroy: true,
        searching: true,
        fixedHeader: true,
        pageLength: 100,
        autoWidth: false,
        progress: false,
        data: [],
        "columns": [
            {"data": "Address"},
            {"data": "NowBlockNum"},
            {"data": "NowBlockHash"},
            {"data": "LastSolidityBlockNum"},
            {"data": "gRPC"},
            {"data": "gRPCMonitor"},
            {"data": "Message"}
        ],

    });

    if (window.WebSocket != undefined) {
        connection = new WebSocket(wsInfoUrl);

        connection.onopen = function (event) {
            console.log("ws on open");
        };

        connection.onclose = function () {
            console.log("ws on close");
        };

        connection.onmessage = function (event) {
            if (event.data == null || event.data === "") {
                return "";
            }

            var resultData = JSON.parse(event.data);

            for (var i = 0; i < resultData.data.length; ++i) {
                var arr = [];
                arr[0] = resultData.data[i].Address;

                arr[1] = resultData.data[i].NowBlockNum;
                arr[2] = "****";
                if (resultData.data[i].NowBlockHash.length !== 0) {
                    arr[2] = resultData.data[i].NowBlockHash.substring(0, 4) + "****" + resultData.data[i].NowBlockHash.substring(resultData.data[i].NowBlockHash.length - 4, resultData.data[i].NowBlockHash.length);
                }

                arr[3] = resultData.data[i].LastSolidityBlockNum;

                if (resultData.data[i].gRPC <= 0) {
                    arr[4] = '<p class="red">0</p>';
                } else if (resultData.data[i].gRPC < 100) {
                    arr[4] = '<p class="green">' + resultData.data[i].gRPC + '</p>';
                } else if (resultData.data[i].gRPC < 300) {
                    arr[4] = '<p class="blue">' + resultData.data[i].gRPC + '</p>';
                } else {
                    arr[4] = '<p style="color: #F39C12;">' + resultData.data[i].gRPC + '</p>';
                }

                arr[5] = "--";
                if (resultData.data[i].gRPCMonitor.length !== 0) {
                    arr[5] = '<span class="sparklines_ping">' + resultData.data[i].gRPCMonitor + '</span>'
                }

                if (resultData.data[i].Message === 'success') {
                    arr[6] = '<p class="green">' + resultData.data[i].Message + '</p>';
                } else {
                    arr[6] = '<p class="red">' + resultData.data[i].Message + '</p>';
                }

                resultData.data[i].Address = arr[0];
                resultData.data[i].NowBlockNum = arr[1];
                resultData.data[i].NowBlockHash = arr[2];
                resultData.data[i].LastSolidityBlockNum = arr[3];
                resultData.data[i].gRPC = arr[4];
                resultData.data[i].gRPCMonitor = arr[5];
                resultData.data[i].Message = arr[6];
            }

            table.rows().remove();
            table.rows.add(resultData.data).draw();

            initPing();
        }
    }

});

function initPing() {
    $('.sparklines_ping').sparkline('html', {
        type: 'bar',
        zeroColor: '#ff0000',
        barColor: '#00bf00',
        colorMap: {
            '1:99':'#1CBD20',
            '100:299': '#48A4DF',
            '300:':'#F39C12'
        },
        //tooltipFormat: $.spformat('{{value}}', 'tooltip-class'),
    });
}

function initTag() {
    axios.get(settingsUrl).then(function (response) {

        if (response == null) {
            return;
        }

        if (response.data == null) {
            return;
        }

        for (var i = 0; i < response.data.length; ++i) {
            var radioStr = `
            <div class="radio">
                <label>
                    <input type="radio" class="flat" name="serverTags" value="` + response.data[i].tag + `">` +
                response.data[i].tag + `
                </label>
            `;

            if (response.data[i].isOpenMonitor === "true") {
                radioStr += `
                <small class="fa fa-bell green">已开启钉钉报警</small>
                `
            } else {
                radioStr += `
                <small class="fa fa-bell">未开启钉钉报警</small>
                `
            }

            radioStr += `
             </div>
            `;
            $("#serverRadios").append(radioStr);
        }

        $(":radio[name='serverTags']:first").attr("checked","true");

        $(":radio[name='serverTags']").change(function () {
            if (connection != undefined) {
                connection.send(this.value);
            }
        });
    }).catch(function (error) {
        console.log(error);
    });
}

function initRunTime() {
    var timestamp = new Date();
    axios.get(runTimeUrl).then(function(response) {
        timestamp.setTime(response.data * 1000);

        setInterval(function () {
            var currentData = new Date();
            $("#runTime").text(getTime(parseInt((currentData - timestamp) / 1000)));
        }, 1000);
    }).catch(function (error) {
        console.log(error);
    });
}

function getTime(seconds) {
    if (seconds <= 60) {
        return seconds + 's';
    } else if (seconds < 3600) {
        return parseInt(seconds / 60) + 'm' + (seconds % 60) + "s";
    } else if (seconds < 86400) {
        return parseInt(seconds / 3600) + 'h' + parseInt((seconds % 3600) / 60) + 'm' + (((seconds % 3600) % 60)) + 's';
    } else {
        return parseInt(seconds / 86400) + 'd' + parseInt(seconds % 86400 / 3600) + 'h' + parseInt((seconds % 86400) % 3600 / 60) + 'm' + (((seconds % 86400) % 3600 % 60)) + 's';
    }

    return seconds + 's';
}

/*
var tag = "MainNetFullNodes";

var infoUrl = serverHost + "/v1/monitor/info/tag/";
var settingsUrl = serverHost + "/v1/monitor/settings/";
var runTimeUrl = serverHost + "/v1/monitor/program-info/";

var table = $('#showdatatable').DataTable({
    destroy: true,
    searching: true,
    fixedHeader: true,
    pageLength: 100,
    autoWidth: false,
    progress: false,
    ajax: {
        url: infoUrl + tag,
        type: "GET",
        dataSrc: function (response) {
            if (response == null) {
                return "";
            }

            for (var i = 0; i < response.data.length; ++i) {
                var arr = [];
                arr[0] = response.data[i].Address;

                arr[1] = response.data[i].NowBlockNum;
                arr[2] = response.data[i].NowBlockHash.substring(0, 4) + "****" + response.data[i].NowBlockHash.substring(response.data[i].NowBlockHash.length - 4, response.data[i].NowBlockHash.length);

                arr[3] = response.data[i].LastSolidityBlockNum;

                if (response.data[i].Ping <= 0) {
                    arr[4] = '<p class="red">0</p>';
                } else if (response.data[i].Ping < 100) {
                    arr[4] = '<p class="green">' + response.data[i].Ping + '</p>';
                } else if (response.data[i].Ping < 300) {
                    arr[4] = '<p class="blue">' + response.data[i].Ping + '</p>';
                } else {
                    arr[4] = '<p style="color: #F39C12;">' + response.data[i].Ping + '</p>';
                }

                arr[5] = "--";
                if (response.data[i].PingMonitor !== '') {
                    arr[5] = '<span class="sparklines_ping">' + response.data[i].PingMonitor + '</span>'
                }

                if (response.data[i].Message === 'success') {
                    arr[6] = '<p class="green">' + response.data[i].Message + '</p>';
                } else {
                    arr[6] = '<p class="red">' + response.data[i].Message + '</p>';
                }

                response.data[i] = arr;
            }
            return response.data;
        }
    }
});

// 页面加载后执行
$(document).ready(function () {
    initTag();

    initRunTime();

    initTable();
});

function initPing() {
    $('.sparklines_ping').sparkline('html', {
        type: 'bar',
        zeroColor: '#ff0000',
        barColor: '#00bf00',
        colorMap: {
            '1:99':'#1CBD20',
            '100:299': '#48A4DF',
            '300:':'#F39C12'
        },
    });
}

function initTag() {
    axios.get(settingsUrl).then(function (response) {

        if (response == null) {
            return;
        }

        if (response.data == null) {
            return;
        }

        for (var i = 0; i < response.data.length; ++i) {
            var radioStr = `
            <div class="radio">
                <label>
                    <input type="radio" class="flat" name="serverTags" value="` + response.data[i].tag + `">` +
                response.data[i].tag + `
                </label>
            `;

            if (response.data[i].isOpenMonitor === "true") {
                radioStr += `
                <small class="fa fa-bell green">已开启钉钉报警</small>
                `
            } else {
                radioStr += `
                <small class="fa fa-bell">未开启钉钉报警</small>
                `
            }

            radioStr += `
             </div>
            `;
            $("#serverRadios").append(radioStr);
        }

        $(":radio[name='serverTags']:first").attr("checked","true");

        $(":radio[name='serverTags']").change(function () {
            tag = this.value;
            table.ajax.url(infoUrl + tag);
            table.ajax.reload(initPing);
        });
    }).catch(function (error) {
        console.log(error);
    });
}

function initRunTime() {
    setInterval(function () {
        axios.get(runTimeUrl).then(function(response) {
            $("#runTime").text(response.data);
        }).catch(function (error) {
            console.log(error);
        })
    }, 1000);
}

function initTable() {
    setInterval(function () {
        table.ajax.reload(initPing);
    }, 3000);
}
*/