/**
 * subscriptionService.ts
 *
 * Abstraction layer for subscription state management.
 * Currently uses a mock implementation. Replace the body of
 * `fetchSubscriptionStatus` with RevenueCat SDK calls when integrating.
 *
 * RevenueCat integration path (Phase 23+):
 *   import Purchases from 'react-native-purchases';
 *   const info = await Purchases.getCustomerInfo();
 *   return { isPro: info.entitlements.active['pro'] !== undefined };
 */

/** Represents the user's current subscription status. */
export interface SubscriptionStatus {
  isPro: boolean;
  /** ISO 8601 expiry date string, or null if not subscribed / unknown. */
  expiresAt: string | null;
  /** RevenueCat product identifier of the active plan, or null. */
  productId: string | null;
}

/** Product identifiers – match RevenueCat dashboard entries exactly. */
export const PRODUCT_IDS = {
  MONTHLY: 'smart_shopper_pro_monthly',
  ANNUAL: 'smart_shopper_pro_annual',
} as const;

export type ProductId = (typeof PRODUCT_IDS)[keyof typeof PRODUCT_IDS];

/**
 * Fetches the current subscription status.
 * Mock: always returns { isPro: false }.
 */
export async function fetchSubscriptionStatus(): Promise<SubscriptionStatus> {
  // TODO (Phase 23): replace with RevenueCat getCustomerInfo()
  await _simulateLatency(300);
  return { isPro: false, expiresAt: null, productId: null };
}

/**
 * Initiates a purchase flow for the given product.
 * Mock: simulates success and returns isPro: true.
 */
export async function purchaseSubscription(
  productId: ProductId,
): Promise<SubscriptionStatus> {
  // TODO (Phase 23): replace with Purchases.purchaseProduct(productId)
  console.log(`[SubscriptionService] Mock purchase initiated: ${productId}`);
  await _simulateLatency(800);
  return {
    isPro: true,
    expiresAt: _mockExpiryDate(productId),
    productId,
  };
}

/**
 * Restores previously purchased subscriptions.
 * Mock: returns isPro: false (no prior purchase in mock).
 */
export async function restorePurchases(): Promise<SubscriptionStatus> {
  // TODO (Phase 23): replace with Purchases.restorePurchases()
  await _simulateLatency(600);
  return { isPro: false, expiresAt: null, productId: null };
}

// ── Private helpers ──────────────────────────────────────────────────────────

function _simulateLatency(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function _mockExpiryDate(productId: ProductId): string {
  const d = new Date();
  if (productId === PRODUCT_IDS.ANNUAL) {
    d.setFullYear(d.getFullYear() + 1);
  } else {
    d.setMonth(d.getMonth() + 1);
  }
  return d.toISOString();
}
