define(["knockout"], function(ko) {
  var utils = utils || {};

  ko.bindingHandlers.modal = {
    init: function(element, valueAccessor) {
      $(element).modal({
        show: false
      });

      var value = valueAccessor();
      if (typeof value === "function") {
        $(element).on("hide.bs.modal", function() {
          value(false);
        });
      }
      ko.utils.domNodeDisposal.addDisposeCallback(element, function() {
        $(element).modal("destroy");
      });
    },
    update: function(element, valueAccessor) {
      var value = valueAccessor();
      if (ko.utils.unwrapObservable(value)) {
        $(element).modal("show");
      } else {
        $(element).modal("hide");
      }
    }
  };

  utils.objectToFormData = function(obj, fd) {
    function isObject(value) {
      return value === Object(value);
    }

    function isArray(value) {
      return Array.isArray(value);
    }

    function isFile(value) {
      return value instanceof File;
    }

    function makeArrayKey(key) {
      if (key.length > 2 && key.lastIndexOf("[]") === key.length - 2) {
        return key;
      } else {
        return key + "[]";
      }
    }

    function objectToFormData(obj, fd, pre) {
      fd = fd || new FormData();

      Object.keys(obj).forEach(function(prop) {
        var key = pre ? pre + "[" + prop + "]" : prop;

        if (isObject(obj[prop]) && !isArray(obj[prop]) && !isFile(obj[prop])) {
          objectToFormData(obj[prop], fd, key);
        } else if (isArray(obj[prop])) {
          obj[prop].forEach(function(value) {
            var arrayKey = makeArrayKey(key);

            if (isObject(value) && !isFile(value)) {
              objectToFormData(value, fd, arrayKey);
            } else {
              fd.append(arrayKey, value);
            }
          });
        } else {
          fd.append(key, obj[prop]);
        }
      });

      return fd;
    }

    return objectToFormData(obj, fd);
  };

  return utils;
});
