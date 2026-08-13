# Component architecture

Shared primitives live in `common`; `squad-profile` owns the localized typed
summary used by member and admin surfaces, while `admin/squad-profile` owns
the editor fields. Member journeys are grouped by feature, `layout` owns
shells, and `admin` owns console-only panels. Components keep data loading in
composables or parent panels and emit validated user intent upward.
