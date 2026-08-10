# Component architecture

Shared primitives live in `common`. Member journeys are grouped by feature, while `layout` owns shells and `admin` owns console-only panels. Components keep data loading in composables or parent panels and emit validated user intent upward.

