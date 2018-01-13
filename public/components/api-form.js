define(["knockout", "toastr", "app/utils", "app/main"], function(
  ko,
  toastr,
  utils,
  webcli
) {
  function AddApiModel(params) {
    var idValue = "";
    var methodValue = "GET";
    var endpointValue = "/";
    var originKey = "";
    if (params && params.context) {
      methodValue = params.context().method;
      endpointValue = params.context().endpoint;
      idValue = params.context().id;

      if (idValue) {
        originKey = params.context().originKey;
      }
    }

    this.originKey = originKey;
    this.method = ko.observable(methodValue);
    this.endpoint = ko.observable(endpointValue);
    this.isFileResult = ko.observable(false);
    this.selectedFile = ko.observable("");
    this.onFileSelected = function(vm, evt) {
      ko.utils.arrayForEach(evt.target.files, function(file) {
        vm.selectedFile(file.name);
      });
    };
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
      var isFileResult = this.isFileResult();
      if (!method) {
        toastr["error"]("Http method required.");
        return;
      }

      if (isFileResult) {
        if (document.getElementById("file").files.length == 0) {
          toastr["error"]("Select a file.");
          return;
        }
      }

      var endpoint = this.endpoint();
      if (!endpoint || (!isFileResult && method == "GET" && endpoint == "/")) {
        toastr["error"]("Endpoint invalid.");
        return;
      }

      var formData = new FormData();
      utils.objectToFormData(JSON.parse(ko.toJSON(this)), formData);

      var file = $("#file");
      if (file.length > 0) {
        var fileInput = file[0].files[0];
        formData.append("file", fileInput);
      }

      var ajaxOptions = {
        type: "POST",
        cache: false,
        data: formData,
        contentType: false,
        processData: false,
        enctype: "multipart/form-data",
        url: "/webcli/api/saveendpoint",
        beforeSend: function() {},
        success: function(data, textStatus, jqXHR) {
          toastr["success"]("Saved...");
          webcli.refreshTree();
        },
        error: function(response) {}
      };

      var jqXHR = ($.ajax(ajaxOptions).always = function(
        data,
        textStatus,
        jqXHR
      ) {});
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
