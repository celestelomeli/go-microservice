### Summary of Changes

Please describe the changes made in the PR. What did you add or modify? Why are these changes necessary?

- Added a new API endpoint for retrieving user details
- Updated the database schema to include a "status" field for users
- Fixed issue with the authentication middleware that caused login failures

### Type of Change

Please indicate the type of change this PR introduces (check all that apply):

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor
- [ ] Documentation update
- [ ] Test update
- [ ] Other (please describe)

### Problem Solved

Explain the problem that was fixed, or the feature that was added. Include any relevant context about why this change is needed.

- The previous login method failed when user status was not active. This change resolves that issue by ensuring the status is checked before authentication.

### Screenshots / GIFs (if applicable)

If this PR includes a visual or UI change, please include screenshots or GIFs to demonstrate the change.

- [Screenshot or GIF showing the UI change]

### Related Issues

- Closes #456 (issue number)
- Follows up on #789

### Testing Instructions

- Describe the tests you ran to verify the changes (manual tests, unit tests, integration tests, etc.)
- Please include specific instructions on how to test the new feature or bug fix.
  - Example: Test the new login flow by attempting to log in with an inactive user.

### Checklist

- [ ] Code is clean and follows project guidelines
- [ ] Unit tests are added for new features
- [ ] Documentation is updated
- [ ] I have run `npm test` (or another testing command) and all tests are passing
- [ ] I have added any relevant information for the reviewer

### Reviewer Checklist

- [ ] I have reviewed the code for quality, clarity, and maintainability
- [ ] I have verified that the change works as expected
- [ ] I have checked for any potential security issues

