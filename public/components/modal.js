define(["knockout"], function(ko) {
  return {
    viewModel: function(params) {
      console.log("PARAMS", params);
      console.log("IS Obs:", ko.isObservable(params.componentName));
      this.params = params || {};
      this.showModal = params.showModal;
      this.title = params.title;
      this.componentName = params.componentName;
    }
  };
});
