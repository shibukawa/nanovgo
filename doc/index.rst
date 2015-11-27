.. NanoVGo documentation master file, created by
   sphinx-quickstart on Fri Nov 27 20:24:55 2015.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

NanoVGo
===================================

.. image:: https://godoc.org/github.com/shibukawa/nanovgo?status.svg
   :target: https://godoc.org/github.com/shibukawa/nanovgo

.. raw:: html

   <iframe src="/demo/" width="900" height="540"></iframe>

`Full Screen </demo/>`_

NanoVGo is a pure golang implementation of `NanoVG <https://github.com/memononen/nanovg>`_. NanoVG is a vector graphics engine inspired by HTML5 Canvas API.

It is depend on cross platform `OpenGL/WebGL library <https://github.com/goxjs/gl>`_. Sample code uses cross platform `glfw wrapper <https://github.com/goxjs/glfw>`_.  I tested on Mac, Windows, and browsers.

.. note::

   To build on gopher.js, it needs `this fix <https://github.com/goxjs/glfw/pull/7>`_ to enable stencil buffer feature now.

.. toctree::
   :maxdepth: 2

   futureplan
   thanks

API Reference
---------------

See `GoDoc <https://godoc.org/github.com/shibukawa/nanovgo>`_

Author
---------------

* `Yoshiki Shibukawa <https://github.com/shibukawa>`_

License
----------

zlib license

Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`

