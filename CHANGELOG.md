# Changelog

## v0.4.0 (2026-06-27)

- Fix backup download (was failing with "file not available")
- Backup filename now includes timestamp
- Add favicon
- Pagination for fill-up and expense lists (10 per page, compact ← 1 … 5 … 18 → navigation)
- Fix CSV import: robust unit stripping (kWh, KWH, Kwh, ltr, gallon, all currencies, case-insensitive)
- Mobile button layout improvements on vehicle detail page

## v0.3.0 (2026-06-27)

- Expense CSV export
- Backup/restore now includes expenses
- Plugin Hybrid vehicle type (supports both L and kWh)
- Bulk delete for fill-ups and expenses
- Floating action button (FAB) on dashboard for quick-add

## v0.2.0 (2026-06-27)

- CSV import for expenses (same flow as fill-ups)
- Single Import button with type selection (Fill-ups or Expenses)
- Dashboard chart shows fuel vs expenses as stacked bars
- Legend shows per-vehicle breakdown (fuel + expenses)
- Auto-create DATA_DIR if it doesn't exist (fixes Unraid startup crash)

## v0.1.0

- Initial release
