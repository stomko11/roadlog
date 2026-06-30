# Changelog

## v0.9.1 (2026-06-30)

- Time-based backup scheduling: pick exact time for daily, day+time for weekly, day-of-month+time for monthly
- Fix: backup failures no longer mark last_run as successful (prevents missed backups)
- Add backup scheduler logging (visible in docker logs)

## v0.9.0 (2026-06-29)

- Automatic backup to WebDAV (Nextcloud) or local path with daily/weekly schedule
- Configurable retention (keep last N backups)
- "Run Now" button for manual trigger
- Rename dashboard chart from "Monthly Spending" to "Spending Overview"
- Remove redundant "Spending by Vehicle" section (info already in stacked chart)

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
