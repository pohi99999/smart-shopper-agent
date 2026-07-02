import Purchases from 'react-native-purchases';
import {
  fetchSubscriptionStatus,
  purchaseSubscription,
  restorePurchases,
  parseCustomerInfo,
  PRODUCT_IDS,
} from './subscriptionService';

jest.mock('react-native-purchases', () => ({
  configure: jest.fn(),
  setLogLevel: jest.fn(),
  getCustomerInfo: jest.fn(),
  purchaseProduct: jest.fn(),
  restorePurchases: jest.fn(),
  LOG_LEVEL: { DEBUG: 'DEBUG' },
}));

describe('subscriptionService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    delete process.env.EXPO_PUBLIC_RC_APPLE_KEY;
    delete process.env.EXPO_PUBLIC_RC_GOOGLE_KEY;
  });

  it('fetchSubscriptionStatus returns initial default status when API key is missing (fallback mode)', async () => {
    const status = await fetchSubscriptionStatus();
    expect(status.isPro).toBe(false);
    expect(status.expiresAt).toBeNull();
    expect(status.productId).toBeNull();
  });

  it('purchaseSubscription returns isPro: true in fallback mode', async () => {
    const status = await purchaseSubscription(PRODUCT_IDS.MONTHLY);
    expect(status.isPro).toBe(true);
    expect(status.productId).toBe(PRODUCT_IDS.MONTHLY);
    expect(status.expiresAt).not.toBeNull();
  });

  it('parseCustomerInfo correctly parses active pro entitlement', () => {
    const mockCustomerInfo: any = {
      entitlements: {
        active: {
          pro: {
            expirationDate: '2027-01-01T00:00:00.000Z',
            productIdentifier: PRODUCT_IDS.MONTHLY,
          },
        },
      },
    };

    const status = parseCustomerInfo(mockCustomerInfo);
    expect(status.isPro).toBe(true);
    expect(status.expiresAt).toBe('2027-01-01T00:00:00.000Z');
    expect(status.productId).toBe(PRODUCT_IDS.MONTHLY);
  });

  it('fetches subscription status from RevenueCat when initialized', async () => {
    process.env.EXPO_PUBLIC_RC_APPLE_KEY = 'appl_mock_key';
    const mockCustomerInfo: any = {
      entitlements: {
        active: {
          pro: {
            expirationDate: '2027-01-01T00:00:00.000Z',
            productIdentifier: PRODUCT_IDS.ANNUAL,
          },
        },
      },
    };

    (Purchases.getCustomerInfo as jest.Mock).mockResolvedValueOnce(mockCustomerInfo);

    const status = await fetchSubscriptionStatus();
    expect(status.isPro).toBe(true);
    expect(status.productId).toBe(PRODUCT_IDS.ANNUAL);
  });

  it('restorePurchases returns status in fallback mode', async () => {
    const status = await restorePurchases();
    expect(status.isPro).toBe(false);
  });
});
