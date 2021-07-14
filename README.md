# Falanx
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-pink.svg)](https://www.gnu.org/licenses/lgpl-3.0)
## Introduction
falanx: a byzantine broadcast protocol

It is a protocol (naive demo) trying to achieve a Byzantine ordered consensus according to 
[CRYPTO2020 Order-Fairness for Byzantine Consensus](https://eprint.iacr.org/2020/269.pdf) and [OSDI2020 Byzantine Ordered Consensus without Byzantine Oligarchy](https://eprint.iacr.org/2020/1300.pdf).

Now, we are trying to implement a *Practical Byzantine Ordered Consensus Protocol*, which could achieve order-fairness and avoid Byzantine behavior as far as possible. We call it *Phalanx*, regular order for block generation. We would like to propose a practical protocol component which could be combined with all kinds of existing BFT algorithm to achieve *order-fairness* and somehow the *correctness* in both synchronous and asynchronous network.

If you are interested in it, please contact us with e-mail (grivn.wang@gmail.com) or propose issues.
