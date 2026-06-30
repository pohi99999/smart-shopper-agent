import React, { useState } from 'react';
import { Modal } from 'react-native';
import { SubscriptionProvider } from './src/context/SubscriptionContext';
import ShoppingListScreen from './src/screens/ShoppingListScreen';
import PaywallScreen from './src/screens/PaywallScreen';

export default function App() {
  const [paywallVisible, setPaywallVisible] = useState(false);

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


