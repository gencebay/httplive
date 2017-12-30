define(["knockout"], function(ko) {
  return {
    viewModel: function(params) {
      console.log("Add API initialized", params);

      this.chosenValue = ko.observable();
      this.like = function() {
        this.chosenValue("like");
      }.bind(this);
      this.dislike = function() {
        this.chosenValue("dislike");
      }.bind(this);
    }
  };
});
