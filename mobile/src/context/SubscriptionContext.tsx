import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  fetchSubscriptionStatus,
  purchaseSubscription,
  restorePurchases,
  type ProductId,
  type SubscriptionStatus,
  PRODUCT_IDS,
} from '../services/subscriptionService';

// ── Types ─────────────────────────────────────────────────────────────────────

interface SubscriptionContextValue {
  /** Whether the user has an active Pro subscription. */
  isPro: boolean;
  /** Full subscription status object. */
  status: SubscriptionStatus;
  /** True while a network/purchase operation is in progress. */
  isLoading: boolean;
  /** Error message from the last failed operation, or null. */
  error: string | null;
  /** Trigger a purchase flow for the given product. */
  subscribe: (productId?: ProductId) => Promise<void>;
  /** Restore previous purchases. */
  restore: () => Promise<void>;
  /** Force-refresh subscription status from the service. */
  refresh: () => Promise<void>;
}

const DEFAULT_STATUS: SubscriptionStatus = {
  isPro: false,
  expiresAt: null,
  productId: null,
};

// ── Context ───────────────────────────────────────────────────────────────────

const SubscriptionContext = createContext<SubscriptionContextValue>({
  isPro: false,
  status: DEFAULT_STATUS,
  isLoading: false,
  error: null,
  subscribe: async () => {},
  restore: async () => {},
  refresh: async () => {},
});

// ── Provider ──────────────────────────────────────────────────────────────────

interface SubscriptionProviderProps {
  children: React.ReactNode;
}

export function SubscriptionProvider({ children }: SubscriptionProviderProps) {
  const [status, setStatus] = useState<SubscriptionStatus>(DEFAULT_STATUS);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const s = await fetchSubscriptionStatus();
      setStatus(s);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch subscription status');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const subscribe = useCallback(
    async (productId: ProductId = PRODUCT_IDS.MONTHLY) => {
      setIsLoading(true);
      setError(null);
      try {
        const s = await purchaseSubscription(productId);
        setStatus(s);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Purchase failed');
      } finally {
        setIsLoading(false);
      }
    },
    [],
  );

  const restore = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const s = await restorePurchases();
      setStatus(s);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Restore failed');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch status on mount
  useEffect(() => {
    refresh();
  }, [refresh]);

  const value = useMemo<SubscriptionContextValue>(
    () => ({
      isPro: status.isPro,
      status,
      isLoading,
      error,
      subscribe,
      restore,
      refresh,
    }),
    [status, isLoading, error, subscribe, restore, refresh],
  );

  return (
    <SubscriptionContext.Provider value={value}>
      {children}
    </SubscriptionContext.Provider>
  );
}

// ── Hook ──────────────────────────────────────────────────────────────────────

/**
 * Consume the global subscription context.
 * Must be used inside <SubscriptionProvider>.
 */
export function useSubscription(): SubscriptionContextValue {
  return useContext(SubscriptionContext);
}

export { PRODUCT_IDS };
export type { ProductId, SubscriptionStatus };
