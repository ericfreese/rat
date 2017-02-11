About
=====

This library implements tries, also known as prefix trees, using minimal acyclic
finite-state automata for the Go programming language (http://golang.org/).

The implementation is based on [Jan Daciuk, Stoyan Mihov, Bruce W. Watson,
Richard E. Watson (2000)](http://goo.gl/0XLPo). "Incremental Construction of
Minimal Acyclic Finite-State Automata". Computational Linguistics: March 2000,
Vol. 26, No. 1, Pages 3-16.

The javascript equivalent of this library can be found at
[MathieuTurcotte/node-trie](https://github.com/MathieuTurcotte/node-trie).

Installing
==========

    $ go get github.com/MathieuTurcotte/go-trie/gtrie

Documentation
=============

Read it [online](http://go.pkgdoc.org/github.com/MathieuTurcotte/go-trie/gtrie) or run

    $ go doc github.com/MathieuTurcotte/go-trie/gtrie

License
=======

This code is free to use under the terms of the [MIT license](http://mturcotte.mit-license.org/).
