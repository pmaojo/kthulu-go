## 2025-12-24 - Parallelized Product Price Loading
**Learning:** SQL `OR` conditions on different columns (`product_id` OR `product_variant_id`) often prevent optimal index usage and force sequential scans or inefficient bitmap unions. Splitting these into parallel queries targets specific indexes and improves concurrency.
**Action:** When fetching related data that can belong to multiple parent types (polymorphic associations), prefer separate parallel queries over a single query with `OR` conditions.
