import React from 'react';
import { render } from '@testing-library/react-native';
import App from './App';

// Mock expo-linking
jest.mock('expo-linking', () => ({
  parse: jest.fn(),
  getInitialURL: jest.fn().mockResolvedValue(null),
  addEventListener: jest.fn().mockReturnValue({ remove: jest.fn() }),
}));

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
jest.mock('./src/hooks/useShoppingOptimizer', () => {
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

describe('App Component', () => {
  it('renders without crashing wrapped in Sentry', () => {
    const { getByText } = render(<App />);
    expect(getByText('Smart Shopper')).toBeTruthy();
  });
});
