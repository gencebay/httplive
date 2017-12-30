define(["knockout"], function(ko) {
  return {
    viewModel: function(params) {
      console.log("DELETE API initialized", params);

      this.flag = ko.observable();
      this.yes = function() {
        this.flag(true);
      }.bind(this);
      this.no = function() {
        this.chosenValue(false);
      }.bind(this);
    }
  };
});
