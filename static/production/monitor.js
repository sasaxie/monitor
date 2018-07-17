var monitorUrl = wsServerHost + "/v1/monitor/ws/tag";
var settingsUrl = serverHost + "/v1/server-group-config/settings/";

var connection;

var gRPCMap = new Map();
var WitnessMissBlockMap = new Map();

function getGRPCOptionBar(k, d) {
    // 指定图表的配置项和数据
    var option = {
        title: {
            text: '过去1天的gRPC监控数据'
        },
        tooltip: {},
        legend: {
            data: ['gRPC']
        },
        xAxis: {
            data: k
        },
        yAxis: {},
        series: [{
            barMinHeight: 1,
            name: 'gRPC',
            type: 'bar',
            data: d
        }],
        dataZoom: [
            {
                type: "slider",
                start: 0,
                end: 100,
            },
            {
                type: "inside",
                start: 0,
                end: 100,
            }
        ],
        visualMap: {
            pieces: [
                {min: 299, label: ">299ms"},
                {min: 99, max: 299, label: ">99ms"},
                {min: 1, max: 99, label: ">0ms"},
                {max: 0, label: "0"}
            ],
            color: ["#F39C12", "#48A4DF", "#1CBD20", "red"],
        },
    };

    return option;
}

function getWitnessOptionLine(k, d) {
    var option = {
        title: {
            text: '过去1天的Witness Miss Block监控数据'
        },
        tooltip: {},
        legend: {
            data: ['Miss Block']
        },
        xAxis: {
            data: k
        },
        yAxis: {},
        series: [{
            name: 'Miss Block',
            type: 'line',
            data: d
        }],
        dataZoom: [
            {
                type: "slider",
                start: 0,
                end: 100,
            },
            {
                type: "inside",
                start: 0,
                end: 100,
            }
        ],
    };
    return option;
}

function initEChartLine() {
    setTimeout(function () {
        for (var v of WitnessMissBlockMap.values()) {
            v.resize();
        }

        for (var v of gRPCMap.values()) {
            v.resize();
        }
    }, 500);
}

// 页面加载后执行
$(document).ready(function () {
    initTag();

    if (window.WebSocket != undefined) {
        connection = new WebSocket(monitorUrl);

        connection.onopen = function (event) {
            console.log("ws on open");
        };

        connection.onclose = function () {
            console.log("ws on close");
            new PNotify({
                title: '断开连接',
                text: '请刷新浏览器重试...',
                hide: false,
                styling: 'bootstrap3'
            });
        };

        connection.onmessage = function (event) {
            if (event == null || event.data === "null") {
                return;
            }

            var jsonAllObj = JSON.parse(event.data);
            var jsonGRPC = jsonAllObj.GRPCResponse;
            var jsonWitnessMissBlock = jsonAllObj.WitnessMissBlockResponse;

            var idCounts = 1;

            for (var key in jsonGRPC) {
                var id = "gRPC" + idCounts;

                var gRPCData = jsonGRPC[key].Data;
                var gRPCCounts = jsonGRPC[key].Date;

                if (typeof(gRPCMap.get(id)) == "undefined") {
                    var obj = `
                    <div class="row">
                        <div class="col-md-12">
                            <div class="x_panel">
                                <div class="x_title">
                                    <h2>[` + idCounts + `/` + Object.keys(jsonGRPC).length + `] ` + key + `</h2>
                                    <div class="clearfix"></div>
                                </div>
                                <div class="x_content">
                                    <div id="` + id + `"
                                         style="height:260px;"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                    `;

                    $("#gRPCContent").append(obj);

                    var option = getGRPCOptionBar(gRPCCounts, gRPCData);
                    var echartBar = echarts.init(document.getElementById(id));
                    echartBar.setOption(option);

                    idCounts++;

                    gRPCMap.set(id, echartBar);
                } else {
                    var echartBar = gRPCMap.get(id);
                    echartBar.setOption({
                        xAxis: {
                            data: gRPCCounts
                        },
                        series: [{
                            barMinHeight: 1,
                            name: 'gRPC',
                            type: 'bar',
                            data: gRPCData
                        }],
                    });

                    idCounts++;
                }
            }

            var idCounts = 1;

            for (var key in jsonWitnessMissBlock) {
                var id = "WitnessMissBlock" + idCounts;

                var data = jsonWitnessMissBlock[key].Data;
                var date = jsonWitnessMissBlock[key].Date;

                if (typeof(WitnessMissBlockMap.get(id)) == "undefined") {
                    var obj = `
                    <div class="row">
                        <div class="col-md-12">
                            <div class="x_panel">
                                <div class="x_title">
                                    <h2>[` + idCounts + `/` + Object.keys(jsonWitnessMissBlock).length + `] ` + key + `</h2>
                                    <div class="clearfix"></div>
                                </div>
                                <div class="x_content">
                                    <div id="` + id + `"
                                         style="height:260px;"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                    `;

                    $("#WitnessMissBlockContent").append(obj);

                    var option = getWitnessOptionLine(date, data);
                    var echartBar = echarts.init(document.getElementById(id));
                    echartBar.setOption(option);

                    idCounts++;

                    WitnessMissBlockMap.set(id, echartBar);
                } else {
                    var echartBar = WitnessMissBlockMap.get(id);
                    echartBar.setOption({
                        xAxis: {
                            data: date
                        },
                        series: [{
                            name: 'Miss Block',
                            type: 'line',
                            data: data
                        }],
                    });

                    idCounts++;
                }
            }

            for (var v of gRPCMap.values()) {
                v.resize();
            }

            for (var v of WitnessMissBlockMap.values()) {
                v.resize();
            }
        }
    }

});

function initTag() {
    $.ajax({
        url: settingsUrl,// 获取自己系统后台用户信息接口
        type: "GET",
        dataType: "json",
        success: function (response) {
            if (response == null) {
                return;
            }

            if (response === "redirect") {
                window.location.href = serverHost + "/static/production/login.html"
            }

            for (var i = 0; i < response.length; ++i) {
                var radioStr = `
            <div class="radio">
                <label>
                    <input type="radio" class="flat" name="serverTags" value="` + response[i].tag + `">
                    &nbsp;` +
                    response[i].tag + `
                </label>
                </div>
            `;

                // if (response[i].isOpenMonitor === "true") {
                //     radioStr += `
                // <small class="fa fa-bell green">已开启钉钉报警</small>
                // `
                // } else {
                //     radioStr += `
                // <small class="fa fa-bell">未开启钉钉报警</small>
                // `
                // }

                $("#serverRadios").append(radioStr);
            }

            $(":radio[name='serverTags']:first").attr("checked", "true");

            $('#serverRadios').iCheck({
                radioClass: 'iradio_flat-green'
            });

            $('#serverRadios input').on('ifChecked', function () {
                if (connection != undefined) {
                    $("#gRPCContent").empty();
                    $("#WitnessMissBlockContent").empty();
                    gRPCMap.clear();
                    WitnessMissBlockMap.clear();
                    connection.send(this.value);
                }
            });
        },

        error: function (response) {
            console.log(response);
        }
    });

}