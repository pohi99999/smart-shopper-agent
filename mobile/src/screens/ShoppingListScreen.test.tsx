import React from 'react';
import { render } from '@testing-library/react-native';

jest.mock('react-native-purchases', () => ({
  configure: jest.fn(),
  setLogLevel: jest.fn(),
  getCustomerInfo: jest.fn(),
  purchaseProduct: jest.fn(),
  restorePurchases: jest.fn(),
  LOG_LEVEL: { DEBUG: 'DEBUG' },
}));

import ShoppingListScreen from './ShoppingListScreen';

// Mock react-native-maps
jest.mock('react-native-maps', () => {
  const React = require('react');
  const MapView = (props: any) => React.createElement('MapView', props, props.children);
  const Marker = (props: any) => React.createElement('Marker', props, props.children);
  return {
    __esModule: true,
    default: MapView,
    Marker,
  };
});

// Mock the hook to avoid actual location requests and api calls
jest.mock('../hooks/useShoppingOptimizer', () => {
  return {
    useShoppingOptimizer: () => ({
      inputText: '',
      setInputText: jest.fn(),
      loading: false,
      result: null,
      coords: null,
      handleOptimize: jest.fn(),
    }),
  };
});

describe('ShoppingListScreen', () => {
  it('renders correctly', () => {
    const { getByText, getByPlaceholderText } = render(<ShoppingListScreen />);
    
    // Check if the title is present
    expect(getByText('Smart Shopper')).toBeTruthy();
    
    // Check if the input is present
    expect(getByPlaceholderText('Írd be a listát szabad szöveggel...')).toBeTruthy();
    
    // Check if the button is present
    expect(getByText('Útvonal Optimalizálása')).toBeTruthy();
  });
});
