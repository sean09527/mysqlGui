# MySQL Manager - Release Checklist

## Pre-Release Preparation

### Version Management
- [ ] Update version in `wails.json` (`info.productVersion`)
- [ ] Update version in `wails.json` (`debianPackage.packageVersion`)
- [ ] Update version in `frontend/package.json`
- [ ] Update version in `README.md`
- [ ] Create/update `CHANGELOG.md` with release notes

### Code Quality
- [ ] All tests passing (`make test`)
- [ ] No linting errors
- [ ] Code reviewed and approved
- [ ] Dependencies updated to latest stable versions
- [ ] Security vulnerabilities addressed

### Documentation
- [ ] README.md is up to date
- [ ] User guide completed
- [ ] API documentation updated (if applicable)
- [ ] Installation instructions verified
- [ ] Troubleshooting guide updated
- [ ] Screenshots updated

### Icons and Assets
- [ ] Application icons generated for all platforms
  - [ ] Windows: `build/windows/icon.ico`
  - [ ] macOS: `build/darwin/icon.icns`
  - [ ] Linux: `build/appicon.png`
- [ ] Installer graphics created (optional)
- [ ] DMG background image created (macOS, optional)

## Building

### Local Builds
- [ ] Clean build environment (`make clean`)
- [ ] Build for Windows (`make build-windows`)
- [ ] Build for macOS (`make build-macos`)
- [ ] Build for Linux (`make build-linux`)
- [ ] All builds completed successfully

### Packaging
- [ ] Windows NSIS installer created (`make package-windows`)
- [ ] macOS DMG created (`make package-macos`)
- [ ] Linux DEB package created (`make package-linux`)
- [ ] All packages created successfully

### Code Signing (Production Only)
- [ ] Windows executable signed
- [ ] macOS app signed and notarized
- [ ] Linux packages signed with GPG (optional)

## Testing

### Functional Testing
- [ ] Test on clean Windows 10 installation
- [ ] Test on clean Windows 11 installation
- [ ] Test on macOS 11 (Big Sur) or later
- [ ] Test on macOS with Apple Silicon (M1/M2)
- [ ] Test on Ubuntu 20.04 LTS
- [ ] Test on Ubuntu 22.04 LTS
- [ ] Test on other Linux distributions (Fedora, Arch, etc.)

### Feature Testing
- [ ] Database connection management
  - [ ] Create connection profile
  - [ ] Edit connection profile
  - [ ] Delete connection profile
  - [ ] Test connection
  - [ ] Connect to database
  - [ ] SSH tunnel connection
- [ ] Database browsing
  - [ ] List databases
  - [ ] List tables
  - [ ] View table row counts
  - [ ] Refresh lists
- [ ] Schema management
  - [ ] View table structure
  - [ ] Create new table
  - [ ] Modify table structure
  - [ ] Delete table
  - [ ] View CREATE TABLE DDL
- [ ] Data management
  - [ ] View table data
  - [ ] Insert rows
  - [ ] Update rows
  - [ ] Delete rows
  - [ ] Filter data
  - [ ] Sort data
  - [ ] Pagination
- [ ] SQL query execution
  - [ ] Execute SELECT queries
  - [ ] Execute INSERT/UPDATE/DELETE
  - [ ] Execute DDL statements
  - [ ] View query results
  - [ ] Query history
- [ ] Schema synchronization
  - [ ] Compare schemas
  - [ ] Generate sync script
  - [ ] Preview sync script
  - [ ] Execute sync script
- [ ] Import/Export
  - [ ] Export to SQL
  - [ ] Export to CSV
  - [ ] Export to JSON
  - [ ] Import from SQL
  - [ ] Import from CSV
  - [ ] Import from JSON

### Database Compatibility
- [ ] MySQL 5.7
- [ ] MySQL 8.0
- [ ] MariaDB 10.2
- [ ] MariaDB 10.5+

### Performance Testing
- [ ] Large database (1000+ tables)
- [ ] Large table (100,000+ rows)
- [ ] Complex queries
- [ ] Schema sync with many differences
- [ ] Large data import/export

### Security Testing
- [ ] Password encryption verified
- [ ] SSH key handling verified
- [ ] SQL injection protection verified
- [ ] No sensitive data in logs
- [ ] Secure connection handling

### Installation Testing
- [ ] Windows installer runs successfully
- [ ] Windows uninstaller works correctly
- [ ] macOS DMG mounts and installs correctly
- [ ] Linux DEB installs successfully
- [ ] Application launches after installation
- [ ] Application icon appears correctly

## Release Artifacts

### Files to Distribute
- [ ] `MySQL-Manager.exe` (Windows standalone)
- [ ] `MySQL-Manager-amd64-installer.exe` (Windows installer)
- [ ] `MySQL-Manager-Installer.dmg` (macOS)
- [ ] `mysql-manager` (Linux standalone)
- [ ] `mysql-manager_1.0.0_amd64.deb` (Linux DEB)
- [ ] `SHA256SUMS` (checksums file)
- [ ] `RELEASE_NOTES.md` (release notes)

### File Verification
- [ ] All files exist
- [ ] File sizes are reasonable
- [ ] SHA256 checksums generated
- [ ] Files are not corrupted
- [ ] Virus scan completed (Windows)

## Distribution

### GitHub Release
- [ ] Create git tag (`git tag -a v1.0.0 -m "Release 1.0.0"`)
- [ ] Push tag to GitHub (`git push origin v1.0.0`)
- [ ] Create GitHub Release
- [ ] Upload all release artifacts
- [ ] Add release notes
- [ ] Mark as latest release
- [ ] Publish release

### Documentation Sites
- [ ] Update project website
- [ ] Update download links
- [ ] Update documentation
- [ ] Update screenshots

### Package Managers (Optional)
- [ ] Submit to Homebrew (macOS)
- [ ] Submit to Chocolatey (Windows)
- [ ] Submit to Snap Store (Linux)
- [ ] Submit to Flathub (Linux)
- [ ] Submit to AUR (Arch Linux)

## Post-Release

### Announcements
- [ ] Announce on project website
- [ ] Announce on social media (Twitter, LinkedIn, etc.)
- [ ] Post on Reddit (r/programming, r/mysql, etc.)
- [ ] Post on Hacker News
- [ ] Post on Product Hunt
- [ ] Send email to mailing list (if applicable)
- [ ] Update project README badges

### Monitoring
- [ ] Monitor GitHub Issues for bug reports
- [ ] Monitor download statistics
- [ ] Monitor user feedback
- [ ] Monitor crash reports (if telemetry enabled)

### Documentation
- [ ] Archive release documentation
- [ ] Update roadmap
- [ ] Plan next release
- [ ] Document lessons learned

## Rollback Plan

If critical issues are discovered:

1. [ ] Remove download links
2. [ ] Mark GitHub Release as pre-release
3. [ ] Post announcement about the issue
4. [ ] Fix the issue
5. [ ] Create hotfix release (e.g., v1.0.1)
6. [ ] Follow release process again

## Version Numbering

Follow Semantic Versioning (semver):
- **Major** (1.0.0): Breaking changes
- **Minor** (1.1.0): New features, backwards compatible
- **Patch** (1.0.1): Bug fixes, backwards compatible

## Release Schedule

Suggested release schedule:
- **Major releases**: Every 6-12 months
- **Minor releases**: Every 1-3 months
- **Patch releases**: As needed for critical bugs

## Support

After release:
- Monitor GitHub Issues
- Respond to user questions
- Provide bug fixes in patch releases
- Collect feedback for next release

## Notes

- Keep this checklist updated with each release
- Document any issues encountered during release
- Improve the process based on lessons learned
- Automate as much as possible (CI/CD)

---

**Release Manager**: _________________

**Release Date**: _________________

**Version**: _________________

**Status**: ☐ In Progress  ☐ Complete  ☐ Rolled Back
