var wsInfoUrl = wsServerHost + "/v1/wsmonitor/tag";
var settingsUrl = serverHost + "/v1/server-group-config/settings/";

var echartLine;
function initEchartBar() {
    var echartBar = echarts.init(document.getElementById('eee'));
    var echartBar2 = echarts.init(document.getElementById('bbb'));

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
            max: "dataMax",
            data: ["13:00:01", "13:01:01", "13:02:01", "13:03:01", "13:04:01", "13:05:01"]
        },
        yAxis: {
            min: "dataMin"
        },
        series: [{
            barMinHeight: 1,
            name: 'gRPC',
            type: 'bar',
            data: [0, 1, 100, 200, 200, 500]
        }],
        dataZoom: [
            {
                type: "slider",
                start: 96,
                end: 100,
            },
            {
                type: "inside",
                start: 96,
                end: 100,
            }
        ],
        visualMap: {
            pieces: [
                {min: 300, label: ">299ms"},
                {min: 100, max: 299, label: ">99ms"},
                {min: 1, max: 99, label: ">0ms"},
                {max: 0, label: "0"}
            ],
            color: ["#F39C12", "#48A4DF", "#1CBD20", "red"],
        },
    };

    // 使用刚指定的配置项和数据显示图表。
    echartBar.setOption(option);
    echartBar2.setOption(option);

    echartLine = echarts.init(document.getElementById('eeee'));

    var option2 = {
        title: {
            text: '过去1天的Witness Miss Block监控数据'
        },
        tooltip: {},
        legend: {
            data: ['Miss Block']
        },
        xAxis: {
            data: ["13:00:01", "13:01:01", "13:02:01", "13:03:01", "13:04:01", "13:05:01"]
        },
        yAxis: {},
        series: [{
            name: 'Miss Block',
            type: 'line',
            data: [0, 0, 1, 3, 4, 8]
        }],
        dataZoom: [
            {
                type: "slider",
                start: 96,
                end: 100,
            },
            {
                type: "inside",
                start: 96,
                end: 100,
            }
        ],
    };

    echartLine.setOption(option2);
}

function initEchartLine() {
    setTimeout(function () {
        echartLine.resize();
    }, 500);
}

// 页面加载后执行
$(document).ready(function () {
    initEchartBar();

    initTag();

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
            {"data": "gRPC"},
            {"data": "WitnessMissBlock"},
        ],

    });

    if (window.WebSocket != undefined) {
        connection = new WebSocket(wsInfoUrl);

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
            if (event.data == null || event.data === "") {
                return "";
            }

            var resultData = JSON.parse(event.data);

            var tableData = [];
            for (var i = 0; i < resultData.data.length; ++i) {
                var arr = [];
                arr[0] = resultData.data[i].Address;

                arr[1] = resultData.data[i].NowBlockNum;
                arr[2] = "****";
                if (resultData.data[i].NowBlockHash.length !== 0) {
                    arr[2] = resultData.data[i].NowBlockHash.substring(0, 4) + "****" + resultData.data[i].NowBlockHash.substring(resultData.data[i].NowBlockHash.length - 4, resultData.data[i].NowBlockHash.length);
                }

                var o = {};
                o.Address = arr[0];
                o.gRPC = arr[1];
                o.WitnessMissBlock = arr[2];
                tableData[i] = o;
            }

            table.rows().remove();
            table.rows.add(tableData).draw();

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
            '1:99': '#1CBD20',
            '100:299': '#48A4DF',
            '300:': '#F39C12'
        },
        //tooltipFormat: $.spformat('{{value}}', 'tooltip-class'),
    });
}

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
                    connection.send(this.value);
                }
            });
        }

        ,
        error: function (response) {
            console.log(response);

        }
    });

}