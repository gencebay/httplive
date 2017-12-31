define(["knockout"], function(ko) {
  return {
    viewModel: function(params) {
      this.params = params || {};
      this.showModal = params.showModal;
      this.title = params.title;
      this.componentName = params.componentName;
    }
  };
});
