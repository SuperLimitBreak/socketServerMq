socketServerMq
==============

Overview
--------
A light-weight, performant websocket server supporting N-many subscribable message channels based on unique, client provided, subscription keys.

###What is it
* socketServerMq is a wrapper for [channelMq](https://github.com/SuperLimitBreak/channelMq)
* socketServerMq exposes channelMq's API via a JSON schema over eiter a tcp or websocket connection.


Intended Use-Case
-----------------
socketServerMq was designed to intergrate into the rest of the SuperLimitBreak live performance tool-chain,
but is generic enough to be useful as a stand-alone websocket messaging system.