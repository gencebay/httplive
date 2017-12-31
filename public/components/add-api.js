define(["knockout"], function(ko) {
  function AddApiModel(params) {
    this.method = ko.observable();
    this.endpoint = ko.observable(params && (params.endpoint || "/"));
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
