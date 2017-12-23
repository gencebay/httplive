define(
  ["jquery", "jqueryui", "bootstrap", "jsoneditor", "config", "jstree"],
  function($, ui, bootstrap, jsoneditor, config) {
    console.log("jQuery is: ", $.fn.jquery);
    console.log("jQueryUI is: ", $.ui.version);
    console.log("jsoneditor is: ", jsoneditor);
    console.log("config is: ", config);
    console.log("jstree is: ", $.jstree.version);
  }
);
