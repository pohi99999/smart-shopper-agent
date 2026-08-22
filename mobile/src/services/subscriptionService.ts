import { Platform } from 'react-native';
import Purchases, { CustomerInfo, LOG_LEVEL, PurchasesError } from 'react-native-purchases';
import * as Sentry from '@sentry/react-native';

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

export const PRO_ENTITLEMENT = 'pro';

let isInitialized = false;

/**
 * Initializes RevenueCat SDK with platform-specific API key.
 */
export function initRevenueCat(): boolean {
  if (isInitialized) {
    return true;
  }

  const apiKey =
    Platform.OS === 'ios'
      ? process.env.EXPO_PUBLIC_RC_APPLE_KEY
      : process.env.EXPO_PUBLIC_RC_GOOGLE_KEY;

  if (!apiKey) {
    if (__DEV__) {
      console.log('[SubscriptionService] RevenueCat API key missing. Running in fallback mode.');
    }
    return false;
  }

  try {
    if (__DEV__) {
      Purchases.setLogLevel(LOG_LEVEL.DEBUG);
    }
    Purchases.configure({ apiKey });
    isInitialized = true;
    return true;
  } catch (error) {
    Sentry.captureException(error, { extra: { context: '[SubscriptionService] Failed to initialize RevenueCat:' } });
    return false;
  }
}

/**
 * Safely extracts SubscriptionStatus from RevenueCat CustomerInfo.
 */
export function parseCustomerInfo(customerInfo?: CustomerInfo | null): SubscriptionStatus {
  if (!customerInfo || !customerInfo.entitlements || !customerInfo.entitlements.active) {
    return {
      isPro: false,
      expiresAt: null,
      productId: null,
    };
  }

  const entitlement =
    customerInfo.entitlements.active[PRO_ENTITLEMENT] ||
    customerInfo.entitlements.active['pro_entitlement'];

  if (entitlement) {
    return {
      isPro: true,
      expiresAt: entitlement.expirationDate ?? null,
      productId: entitlement.productIdentifier ?? null,
    };
  }

  return {
    isPro: false,
    expiresAt: null,
    productId: null,
  };
}

/**
 * Fetches the current subscription status from RevenueCat.
 */
export async function fetchSubscriptionStatus(): Promise<SubscriptionStatus> {
  const initialized = initRevenueCat();
  if (!initialized) {
    return { isPro: false, expiresAt: null, productId: null };
  }

  try {
    const customerInfo = await Purchases.getCustomerInfo();
    return parseCustomerInfo(customerInfo);
  } catch (error) {
    Sentry.captureException(error, { extra: { context: '[SubscriptionService] Error fetching subscription status:' } });
    return { isPro: false, expiresAt: null, productId: null };
  }
}

/**
 * Initiates a purchase flow for the given product identifier.
 */
export async function purchaseSubscription(
  productId: ProductId,
): Promise<SubscriptionStatus> {
  const initialized = initRevenueCat();
  if (!initialized) {
    return {
      isPro: true,
      expiresAt: _mockExpiryDate(productId),
      productId,
    };
  }

  try {
    const { customerInfo } = await Purchases.purchaseProduct(productId);
    return parseCustomerInfo(customerInfo);
  } catch (error) {
    const purchaseError = error as PurchasesError;
    if (purchaseError?.userCancelled) {
      if (__DEV__) {
        console.log('[SubscriptionService] User cancelled purchase flow');
      }
    } else {
      Sentry.captureException(error, { extra: { context: '[SubscriptionService] Purchase error:' } });
    }
    throw error;
  }
}

/**
 * Restores previously purchased subscriptions.
 */
export async function restorePurchases(): Promise<SubscriptionStatus> {
  const initialized = initRevenueCat();
  if (!initialized) {
    return { isPro: false, expiresAt: null, productId: null };
  }

  try {
    const customerInfo = await Purchases.restorePurchases();
    return parseCustomerInfo(customerInfo);
  } catch (error) {
    Sentry.captureException(error, { extra: { context: '[SubscriptionService] Restore purchases error:' } });
    throw error;
  }
}

// ── Private helpers ──────────────────────────────────────────────────────────

function _mockExpiryDate(productId: ProductId): string {
  const d = new Date();
  if (productId === PRODUCT_IDS.ANNUAL) {
    d.setFullYear(d.getFullYear() + 1);
  } else {
    d.setMonth(d.getMonth() + 1);
  }
  return d.toISOString();
}
