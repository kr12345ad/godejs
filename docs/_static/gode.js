(function () {
  "use strict";

  function ready(fn) {
    if (document.readyState !== "loading") {
      fn();
    } else {
      document.addEventListener("DOMContentLoaded", fn);
    }
  }

  function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
      return navigator.clipboard.writeText(text);
    }
    var area = document.createElement("textarea");
    area.value = text;
    area.setAttribute("readonly", "");
    area.style.position = "fixed";
    area.style.opacity = "0";
    document.body.appendChild(area);
    area.select();
    var copied = false;
    try {
      copied = document.execCommand("copy");
    } catch (error) {
      copied = false;
    }
    document.body.removeChild(area);
    return copied ? Promise.resolve() : Promise.reject(new Error("copy failed"));
  }

  ready(function () {
    var nav = document.getElementById("site-nav");
    var toggle = document.querySelector(".nav-toggle");
    if (nav && toggle) {
      toggle.addEventListener("click", function () {
        var open = nav.classList.toggle("is-open");
        toggle.setAttribute("aria-expanded", open ? "true" : "false");
      });
      nav.addEventListener("click", function (event) {
        if (event.target.closest("a")) {
          nav.classList.remove("is-open");
          toggle.setAttribute("aria-expanded", "false");
        }
      });
      document.addEventListener("keydown", function (event) {
        if (event.key === "Escape" && nav.classList.contains("is-open")) {
          nav.classList.remove("is-open");
          toggle.setAttribute("aria-expanded", "false");
          toggle.focus();
        }
      });
    }

    var filter = document.getElementById("api-filter");
    var cards = Array.prototype.slice.call(document.querySelectorAll(".api-card"));
    var sections = Array.prototype.slice.call(document.querySelectorAll(".api-section"));
    var count = document.getElementById("api-count");
    var empty = document.getElementById("api-empty");
    if (filter && cards.length) {
      var applyFilter = function () {
        var query = filter.value.trim().toLowerCase();
        var visible = 0;
        cards.forEach(function (card) {
          var haystack = (card.getAttribute("data-search") || card.textContent).toLowerCase();
          var match = !query || haystack.indexOf(query) !== -1;
          card.hidden = !match;
          if (match) {
            visible += 1;
          }
        });
        sections.forEach(function (section) {
          section.hidden = !section.querySelector(".api-card:not([hidden])");
        });
        if (count) {
          count.textContent = visible + (visible === 1 ? " module" : " modules");
        }
        if (empty) {
          empty.hidden = visible !== 0;
        }
      };
      filter.addEventListener("input", applyFilter);
      applyFilter();
    }

    Array.prototype.slice.call(document.querySelectorAll("div.highlight")).forEach(function (block) {
      var pre = block.querySelector("pre");
      if (!pre) {
        return;
      }
      var button = document.createElement("button");
      button.type = "button";
      button.className = "copy-code";
      button.textContent = "Copy";
      button.setAttribute("aria-label", "Copy code to clipboard");
      button.addEventListener("click", function () {
        copyText(pre.innerText).then(function () {
          button.textContent = "Copied";
          window.setTimeout(function () { button.textContent = "Copy"; }, 1600);
        }).catch(function () {
          button.textContent = "Select";
        });
      });
      block.appendChild(button);
    });
  });
})();
