var search = new Vue({
  el: '#search',
  data: {
    fullNodesInfo: []
  },

  methods: {
    submitFullNodes: function () {
      var fullNodesStr = $("input[id='tags_1']").val();
      var fullNodes = fullNodesStr.split(",");
      var addressesObject = new Object();
      addressesObject["addresses"] = fullNodes;

      var addressesJson = JSON.stringify(addressesObject);

      var apiUrl = "http://47.254.37.251:8080/v1/monitor/info";
      axios.post(apiUrl, addressesJson).then(function(response) {
        search.$set(search.$data, "fullNodesInfo", response.data.Results);
      })
    },
  },

});