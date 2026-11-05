project = "GodeJS"
copyright = "2026, GodeJS"
author = "GodeJS"

extensions = []
templates_path = ["_templates"]
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store"]

root_doc = "index"
source_suffix = {".rst": "restructuredtext"}

html_theme = "alabaster"
html_title = "GodeJS Documentation"
html_theme_options = {
    "description": "Node.js-compatible runtime on JavaScriptCore",
    "github_button": False,
    "show_powered_by": False,
    "fixed_sidebar": False,
    "show_relbars": False,
}
html_sidebars = {
    "index": [],
    "usage/index": [],
    "usage/api/index": [],
    "**": ["docs-nav.html", "localtoc.html"],
}
html_last_updated_fmt = "%Y-%m-%d"
html_use_index = True
html_split_index = False
html_search_language = "en"
html_static_path = ["_static"]
html_css_files = ["gode.css"]
html_js_files = ["gode.js"]

pygments_style = "sphinx"
