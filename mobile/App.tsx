import React, { useEffect, useState } from 'react';
import { Modal } from 'react-native';
import * as Linking from 'expo-linking';
import * as Sentry from '@sentry/react-native';
import './src/i18n/i18n';
import { SubscriptionProvider } from './src/context/SubscriptionContext';
import ShoppingListScreen from './src/screens/ShoppingListScreen';
import PaywallScreen from './src/screens/PaywallScreen';

Sentry.init({
  dsn: process.env.EXPO_PUBLIC_SENTRY_DSN,
  enableInExpoDevelopment: true,
  debug: __DEV__,
});

function App() {
  const [paywallVisible, setPaywallVisible] = useState(false);

  useEffect(() => {
    const handleDeepLink = (event: { url: string }) => {
      const data = Linking.parse(event.url);
      if (data.path === 'paywall' || data.hostname === 'paywall') {
        setPaywallVisible(true);
      }
    };

    Linking.getInitialURL().then((url) => {
      if (url) {
        handleDeepLink({ url });
      }
    });

    const subscription = Linking.addEventListener('url', handleDeepLink);
    return () => subscription.remove();
  }, []);

  return (
    <SubscriptionProvider>
      <ShoppingListScreen onShowPaywall={() => setPaywallVisible(true)} />
      <Modal
        visible={paywallVisible}
        animationType="slide"
        presentationStyle="pageSheet"
        onRequestClose={() => setPaywallVisible(false)}
      >
        <PaywallScreen onClose={() => setPaywallVisible(false)} />
      </Modal>
    </SubscriptionProvider>
  );
}

export default Sentry.wrap(App);
