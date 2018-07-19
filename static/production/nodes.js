var settingsUrl = serverHost + "/v1/server-group-config/settings/";
var nodesUrl = serverHost + "/v1/node/nodes/tag/";

var tag = "MainNetFullNodes";
var table;

// 页面加载后执行
$(document).ready(function () {
    initTag();

    table = $('#nodesTable').DataTable({
        destroy: true,
        searching: true,
        fixedHeader: true,
        pageLength: 100,
        autoWidth: false,
        progress: false,
        ajax: {
            url: nodesUrl + tag,
            type: "GET",
            dataSrc: function (response) {
                if (response == null) {
                    return "";
                }

                if (response === "not data found, please try again...") {
                    return "";
                }

                $("#nodesCount").text(response.length);

                var res = [];
                for ( var i = 0; i < response.length; i++) {
                    var obj = {};
                    obj.Address = response[i];
                    res.push(obj);
                }

                return res;
            }
        },
        "columns": [
            {"data": "Address"}
            ]
    });
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

            $(":radio[name='serverTags']:first").attr("checked", "true");

            $('#serverRadios').iCheck({
                radioClass: 'iradio_flat-green'
            });

            $('#serverRadios input').on('ifChecked', function () {
                tag = this.value;
                table.ajax.url(nodesUrl + tag);
                table.ajax.reload();
            });
        },
        error: function (response) {
            console.log(response);
        }
    });

}
