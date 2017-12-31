define(["knockout"], function(ko) {
  function AddApiModel(params) {
    var idValue = "";
    var methodValue = "GET";
    var endpointValue = "/";
    if (params && params.context) {
      methodValue = params.context().method;
      endpointValue = params.context().endpoint;
      idValue = params.context().id;
    }

    console.log({ idValue, methodValue, endpointValue });

    this.id = ko.observable(idValue);
    this.method = ko.observable(methodValue);
    this.endpoint = ko.observable(endpointValue);
    this.submit = function() {
      console.log(ko.toJSON(this));
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
