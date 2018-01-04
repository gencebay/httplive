define(["knockout", "toastr", "app/main"], function(ko, toastr, webcli) {
  function AddApiModel(params) {
    var idValue = "";
    var methodValue = "GET";
    var endpointValue = "/";
    var key = "";
    if (params && params.context) {
      methodValue = params.context().method;
      endpointValue = params.context().endpoint;
      idValue = params.context().id;

      if (endpointValue != "/") {
        key = methodValue + endpointValue;
      }
    }

    console.log({ idValue, methodValue, endpointValue });

    this.key = key;
    this.method = ko.observable(methodValue);
    this.endpoint = ko.observable(endpointValue);
    this.methodLabel = ko.computed(function() {
      var method = this.method();
      switch (method) {
        case "GET":
          return "label label-primary";
        case "POST":
          return "label label-success";
        case "PUT":
          return "label label-warning";
        case "DELETE":
          return "label label-danger";
        default:
          break;
      }

      return "Http Live:" + this.port();
    }, this);

    this.submit = function() {
      var method = this.method();
      if (!method) {
        toastr["error"]("Http method required.");
        return;
      }

      var endpoint = this.endpoint();
      if (!endpoint || (method == "GET" && endpoint == "/")) {
        toastr["error"]("Endpoint invalid.");
        return;
      }

      var jqXHR = ($.ajax({
        type: "POST",
        url: "/webcli/api/saveendpoint",
        data: ko.toJSON(this),
        contentType: "application/json; charset=utf-8",
        beforeSend: function() {},
        success: function(data, textStatus, jqXHR) {
          toastr["success"]("Saved...");
          webcli.refreshTree();
        },
        error: function(response) {}
      }).always = function(data, textStatus, jqXHR) {});
    }.bind(this);
  }

  AddApiModel.prototype.dispose = function() {
    // noop
  };

  return {
    viewModel: {
      createViewModel: function(params, componentInfo) {
        return new AddApiModel(params);
      }
    }
  };
});
