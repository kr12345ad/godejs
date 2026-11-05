project = "GodeJS"
copyright = "2026, GodeJS"
author = "GodeJS"

extensions = []
templates_path = []
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
}
html_sidebars = {
    "**": ["about.html", "navigation.html", "searchbox.html"],
}
html_last_updated_fmt = "%Y-%m-%d"
html_use_index = True
html_split_index = False
html_search_language = "en"
html_static_path = ["_static"]
html_css_files = ["dark.css"]

pygments_style = "sphinx"
