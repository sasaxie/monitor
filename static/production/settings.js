var serverConfigUrl = serverHost + "/v1/monitor/server-config";

var editor;

// 页面加载后执行
$(document).ready(function () {
    initServerConfig();
});

function saveServerConfig() {
    var serverConfigJson = editor.getValue();

    axios.post(serverConfigUrl, serverConfigJson).then(function (response) {
        if (response == null ) {
            return
        }
        if (response.data === "redirect") {
            window.location.href = serverHost + "/static/production/login.html"
        }

        if (response.data !== "success") {
            new PNotify({
                title: '保存失败',
                type: 'error',
                styling: 'bootstrap3'
            });
        } else {
            new PNotify({
                title: '保存成功',
                type: 'success',
                styling: 'bootstrap3'
            });
        }
    }).catch(function (error) {
        console.log(error);
    })
}

function initServerConfig() {
    var element = document.getElementById('editor_holder');

    editor = new JSONEditor(element, {
        schema: {
            "title": "服务器配置",
            "type": "object",
            "required": [
                "servers",
            ],
            "properties": {
                "servers": {
                    "type": "array",
                    "title": "服务器组列表",
                    "uniqueItems": true,
                    "items": {
                        "type": "object",
                        "title": "服务器组",
                        "properties": {
                            "setting": {
                                "type": "object",
                                "title": "设置",
                                "properties": {
                                    "isOpenMonitor": {
                                        "type":"string",
                                        "title":"是否开启钉钉监控",
                                        "enum": [
                                          "true",
                                          "false"
                                        ],
                                        "default":"false"
                                    },
                                    "tag": {
                                        "type":"string",
                                        "title":"分组标签"
                                    }
                                }
                            },
                            "addresses": {
                                "type": "array",
                                "title":"地址列表",
                                "format":"table",
                                "items": {
                                    "type":"object",
                                    "title":"地址",
                                    "properties": {
                                        "ip":{
                                            "type":"string",
                                            "title":"IP"
                                        },
                                        "port": {
                                            "type":"number",
                                            "title":"端口号",
                                            "default": 50051
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        iconlib: "fontawesome4",
        theme: "bootstrap3"
    });

    axios.get(serverConfigUrl).then(function (response) {
        if (response == null || response.data ==="") {
            return;
        }

        if (response.data === "redirect") {
            window.location.href = serverHost + "/static/production/login.html"
        }

        var json = response.data;

        editor.setValue(json);
    }).catch(function (error) {
        console.log(error);
    });
}