$(document).ready(function () {
    showData();
});

function recoverRefresh() {
    $('#refresh').removeAttr("disabled");
    $('#refresh').attr("onclick", "refresh()");
}

function refresh() {
    $('#refresh').attr("disabled", "true");
    $('#refresh').removeAttr("onclick");

    setTimeout("recoverRefresh()", 3000);

    showData();
}

function showData() {
    $('#showdatatable').DataTable({
        destroy: true,
        searching: true,
        fixedHeader: true,
        pageLength: 100,
        autoWidth: false,
        progress: false,
        ajax: {
            url: "http://127.0.0.1:8080/v1/monitor/info",
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
                        arr[4] = "--";
                    } else if (response.data[i].Ping < 100) {
                        arr[4] = '<p class="green">' + response.data[i].Ping + '</p>';
                    } else if (response.data[i].Ping < 300) {
                        arr[4] = '<p class="blue">' + response.data[i].Ping + '</p>';
                    } else {
                        arr[4] = '<p style="color: #F39C12;">' + response.data[i].Ping + '</p>';
                    }

                    if (response.data[i].Message === 'success') {
                        arr[5] = '<p class="green">' + response.data[i].Message + '</p>';
                    } else {
                        arr[5] = '<p class="red">' + response.data[i].Message + '</p>';
                    }

                    response.data[i] = arr;
                }
                return response.data;
            }
        }
    });
}