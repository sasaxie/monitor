var wsInfoUrl = wsServerHost + "/v1/wsmonitor/tag";
var settingsUrl = serverHost + "/v1/server-group-config/settings/";
var runTimeUrl = serverHost + "/v1/program/";

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
            {"data": "TotalTransaction"},
            {"data": "gRPC"},
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

            for (var i = 0; i < resultData.data.length; ++i) {
                var arr = [];
                arr[0] = resultData.data[i].Address;

                arr[1] = resultData.data[i].NowBlockNum;
                arr[2] = "****";
                if (resultData.data[i].NowBlockHash.length !== 0) {
                    arr[2] = resultData.data[i].NowBlockHash.substring(0, 4) + "****" + resultData.data[i].NowBlockHash.substring(resultData.data[i].NowBlockHash.length - 4, resultData.data[i].NowBlockHash.length);
                }

                arr[3] = resultData.data[i].LastSolidityBlockNum;
                arr[4] =resultData.data[i].TotalTransaction;

                if (resultData.data[i].gRPC <= 0) {
                    arr[5] = '<p class="red">0</p>';
                } else if (resultData.data[i].gRPC < 100) {
                    arr[5] = '<p class="green">' + resultData.data[i].gRPC + '</p>';
                } else if (resultData.data[i].gRPC < 300) {
                    arr[5] = '<p class="blue">' + resultData.data[i].gRPC + '</p>';
                } else {
                    arr[5] = '<p style="color: #F39C12;">' + resultData.data[i].gRPC + '</p>';
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
                resultData.data[i].TotalTransaction = arr[4];
                resultData.data[i].gRPC = arr[5];
                resultData.data[i].Message = arr[6];

            }

            table.rows().remove();
            table.rows.add(resultData.data).draw();

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

                $("#serverRadios").append(radioStr);
            }

            $(":radio[name='serverTags']:first").attr("checked","true");

            $('#serverRadios').iCheck({
                radioClass: 'iradio_flat-green'
            });

            $('#serverRadios input').on('ifChecked', function () {
                if (connection != undefined) {
                    connection.send(this.value);
                }
            });
        },
        error: function (response) {
            console.log(response);
        }
    });

}

function initRunTime() {
    $.ajax({
        url: runTimeUrl,
        type: "GET",
        dataType: "json",
        success: function (response) {
            if (response === "redirect") {
                window.location.href = serverHost + "/static/production/login.html"
            }

            var timestamp = new Date();
            timestamp.setTime(response * 1000);
            setInterval(function () {
                var currentData = new Date();
                var timeText = getTime(parseInt((currentData - timestamp) / 1000));
                $("#runTime").text(timeText);
            }, 1000);
        }
        ,
        error: function (response) {
            console.log(response);
        }
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
