Future Plan
---------------

Now, it includes all features of `NanoVG <https://github.com/memononen/nanovg>`_.

* Change TrueType font library from `TheOnly92 <https://github.com/TheOnly92/fontstash.go>`_'s code to `pure go freetype <https://github.com/golang/freetype>`_ to support TTC file.
* Add font fallback mechanism.
* Add default font lists for common operating systems.
* Add any path/render image caching system.
* Auto antialias (if device pixel ratio is bigger than 1, turn off AA for performance)
* Add backend for use mobile/gl package.
* Use float64 instead of float32 on gopher.js (see `performance tips <http://www.gopherjs.org/>`_)
