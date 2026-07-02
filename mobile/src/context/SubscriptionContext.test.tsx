import React from 'react';
import { render, act, waitFor, fireEvent } from '@testing-library/react-native';
import { Text, Button } from 'react-native';

jest.mock('react-native-purchases', () => ({
  configure: jest.fn(),
  setLogLevel: jest.fn(),
  getCustomerInfo: jest.fn(),
  purchaseProduct: jest.fn(),
  restorePurchases: jest.fn(),
  LOG_LEVEL: { DEBUG: 'DEBUG' },
}));

import { SubscriptionProvider, useSubscription } from './SubscriptionContext';
import * as subscriptionService from '../services/subscriptionService';

jest.mock('../services/subscriptionService', () => {
  const original = jest.requireActual('../services/subscriptionService');
  return {
    ...original,
    fetchSubscriptionStatus: jest.fn(),
    purchaseSubscription: jest.fn(),
    restorePurchases: jest.fn(),
  };
});

function TestConsumer() {
  const { isPro, isLoading, error, subscribe, restore, refresh } = useSubscription();

  return (
    <>
      <Text testID="isPro">{isPro ? 'Pro' : 'Free'}</Text>
      <Text testID="loading">{isLoading ? 'Loading' : 'Idle'}</Text>
      <Text testID="error">{error || 'None'}</Text>
      <Button title="Subscribe" onPress={() => subscribe(subscriptionService.PRODUCT_IDS.MONTHLY)} />
      <Button title="Restore" onPress={() => restore()} />
      <Button title="Refresh" onPress={() => refresh()} />
    </>
  );
}

describe('SubscriptionContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (subscriptionService.fetchSubscriptionStatus as jest.Mock).mockResolvedValue({
      isPro: false,
      expiresAt: null,
      productId: null,
    });
  });

  it('fetches subscription status on mount', async () => {
    const { getByTestId } = render(
      <SubscriptionProvider>
        <TestConsumer />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByTestId('isPro').children[0]).toBe('Free');
    });
    expect(subscriptionService.fetchSubscriptionStatus).toHaveBeenCalledTimes(1);
  });

  it('handles successful subscription', async () => {
    (subscriptionService.purchaseSubscription as jest.Mock).mockResolvedValueOnce({
      isPro: true,
      expiresAt: '2027-01-01T00:00:00.000Z',
      productId: subscriptionService.PRODUCT_IDS.MONTHLY,
    });

    const { getByTestId, getByText } = render(
      <SubscriptionProvider>
        <TestConsumer />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByTestId('isPro').children[0]).toBe('Free');
    });

    fireEvent.press(getByText('Subscribe'));

    await waitFor(() => {
      expect(getByTestId('isPro').children[0]).toBe('Pro');
    });
  });

  it('handles restore purchases', async () => {
    (subscriptionService.restorePurchases as jest.Mock).mockResolvedValueOnce({
      isPro: true,
      expiresAt: '2027-01-01T00:00:00.000Z',
      productId: subscriptionService.PRODUCT_IDS.ANNUAL,
    });

    const { getByTestId, getByText } = render(
      <SubscriptionProvider>
        <TestConsumer />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByTestId('isPro').children[0]).toBe('Free');
    });

    fireEvent.press(getByText('Restore'));

    await waitFor(() => {
      expect(getByTestId('isPro').children[0]).toBe('Pro');
    });
  });
});
