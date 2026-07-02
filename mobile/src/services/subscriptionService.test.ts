import {
  fetchSubscriptionStatus,
  purchaseSubscription,
  restorePurchases,
  PRODUCT_IDS,
} from './subscriptionService';

describe('subscriptionService', () => {
  it('fetchSubscriptionStatus returns initial default mock status (isPro: false)', async () => {
    const status = await fetchSubscriptionStatus();
    expect(status.isPro).toBe(false);
    expect(status.expiresAt).toBeNull();
    expect(status.productId).toBeNull();
  });

  it('purchaseSubscription returns isPro: true and expiry date for monthly product', async () => {
    const status = await purchaseSubscription(PRODUCT_IDS.MONTHLY);
    expect(status.isPro).toBe(true);
    expect(status.productId).toBe(PRODUCT_IDS.MONTHLY);
    expect(status.expiresAt).not.toBeNull();
  });

  it('purchaseSubscription returns isPro: true and expiry date for annual product', async () => {
    const status = await purchaseSubscription(PRODUCT_IDS.ANNUAL);
    expect(status.isPro).toBe(true);
    expect(status.productId).toBe(PRODUCT_IDS.ANNUAL);
    expect(status.expiresAt).not.toBeNull();
  });

  it('restorePurchases returns status object', async () => {
    const status = await restorePurchases();
    expect(status.isPro).toBe(false);
    expect(status.expiresAt).toBeNull();
    expect(status.productId).toBeNull();
  });
});
