🎯 **What:** Updated the API client fallback URL to use `https://localhost:8080` instead of `http://localhost:8080`, and updated corresponding tests.
⚠️ **Risk:** The use of plaintext `http://` could result in man-in-the-middle attacks where data transmitted between the mobile app and the backend API can be intercepted or altered.
🛡️ **Solution:** Changed the default URL protocol from `http` to `https` in `mobile/src/services/api.ts` and updated the unit tests in `mobile/src/services/api.test.ts`.
