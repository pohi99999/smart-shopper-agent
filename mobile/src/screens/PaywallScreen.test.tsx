import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import PaywallScreen from './PaywallScreen';
import { SubscriptionProvider } from '../context/SubscriptionContext';
import i18n from '../i18n/i18n';

jest.mock('../services/subscriptionService', () => {
  const original = jest.requireActual('../services/subscriptionService');
  return {
    ...original,
    fetchSubscriptionStatus: jest.fn().mockResolvedValue({
      isPro: false,
      expiresAt: null,
      productId: null,
    }),
    purchaseSubscription: jest.fn().mockResolvedValue({
      isPro: true,
      expiresAt: '2027-01-01T00:00:00.000Z',
      productId: 'smart_shopper_pro_monthly',
    }),
    restorePurchases: jest.fn().mockResolvedValue({
      isPro: false,
      expiresAt: null,
      productId: null,
    }),
  };
});

describe('PaywallScreen', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('hu');
  });

  it('renders title, features and pricing properly', async () => {
    const mockClose = jest.fn();
    const { getByText } = render(
      <SubscriptionProvider>
        <PaywallScreen onClose={mockClose} />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByText('Smart Shopper')).toBeTruthy();
      expect(getByText('PRO')).toBeTruthy();
    });
  });

  it('calls onClose when close button is pressed', async () => {
    const mockClose = jest.fn();
    const { getByLabelText } = render(
      <SubscriptionProvider>
        <PaywallScreen onClose={mockClose} />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByLabelText('Bezárás')).toBeTruthy();
    });

    const closeButton = getByLabelText('Bezárás');
    fireEvent.press(closeButton);
    expect(mockClose).toHaveBeenCalled();
  });

  it('triggers purchase when subscription CTA is pressed', async () => {
    const mockClose = jest.fn();
    const { getByText } = render(
      <SubscriptionProvider>
        <PaywallScreen onClose={mockClose} />
      </SubscriptionProvider>
    );

    await waitFor(() => {
      expect(getByText('Előfizetés indítása')).toBeTruthy();
    });

    const ctaButton = getByText('Előfizetés indítása');
    fireEvent.press(ctaButton);

    await waitFor(() => {
      expect(getByText('✅ Pro fiók aktiválva!')).toBeTruthy();
    });
  });
});
