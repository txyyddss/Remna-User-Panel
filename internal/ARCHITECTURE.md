# Internal architecture

Domain packages own validation and workflows, `integrations` owns upstream wire contracts, and `platform` owns reusable runtime mechanics. `app` composes those packages, while `httpapi` exposes the signed API boundary. Dependencies point toward domain contracts and never from a domain into HTTP handlers.
