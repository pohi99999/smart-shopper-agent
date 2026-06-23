import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import ShoppingListScreen from './ShoppingListScreen';
import { useShoppingOptimizer } from '../hooks/useShoppingOptimizer';

// Mock the hook and react-native-maps
jest.mock('../hooks/useShoppingOptimizer');
jest.mock('react-native-maps', () => {
  const { View } = require('react-native');
  const MockMapView = (props: any) => <View testID="map-view" {...props} />;
  const MockMarker = (props: any) => <View testID="map-marker" {...props} />;
  return {
    __esModule: true,
    default: MockMapView,
    Marker: MockMarker,
  };
});

describe('ShoppingListScreen', () => {
  beforeEach(() => {
    (useShoppingOptimizer as jest.Mock).mockReturnValue({
      inputText: '',
      setInputText: jest.fn(),
      loading: false,
      result: null,
      coords: null,
      handleOptimize: jest.fn(),
    });
  });

  it('renders correctly with input field and button', () => {
    const { getByText, getByPlaceholderText } = render(<ShoppingListScreen />);

    // Assert that titles exist
    expect(getByText('Smart Shopper')).toBeTruthy();
    expect(getByText('Személyes bevásárló asszisztens')).toBeTruthy();
    expect(getByText('Mit szeretnél vásárolni?')).toBeTruthy();

    // Assert input field exists
    const input = getByPlaceholderText('Írd be a listát szabad szöveggel...');
    expect(input).toBeTruthy();

    // Assert button exists
    const button = getByText('Útvonal Optimalizálása');
    expect(button).toBeTruthy();
  });

  it('calls setInputText when typing', () => {
    const setInputTextMock = jest.fn();
    (useShoppingOptimizer as jest.Mock).mockReturnValue({
      inputText: '',
      setInputText: setInputTextMock,
      loading: false,
      result: null,
      coords: null,
      handleOptimize: jest.fn(),
    });

    const { getByPlaceholderText } = render(<ShoppingListScreen />);
    const input = getByPlaceholderText('Írd be a listát szabad szöveggel...');

    fireEvent.changeText(input, 'tej, kenyer');
    expect(setInputTextMock).toHaveBeenCalledWith('tej, kenyer');
  });

  it('calls handleOptimize when button is pressed', () => {
    const handleOptimizeMock = jest.fn();
    (useShoppingOptimizer as jest.Mock).mockReturnValue({
      inputText: 'tej',
      setInputText: jest.fn(),
      loading: false,
      result: null,
      coords: null,
      handleOptimize: handleOptimizeMock,
    });

    const { getByText } = render(<ShoppingListScreen />);
    const button = getByText('Útvonal Optimalizálása');

    fireEvent.press(button);
    expect(handleOptimizeMock).toHaveBeenCalled();
  });

  it('shows loading indicator when loading is true', () => {
    (useShoppingOptimizer as jest.Mock).mockReturnValue({
      inputText: 'tej',
      setInputText: jest.fn(),
      loading: true,
      result: null,
      coords: null,
      handleOptimize: jest.fn(),
    });

    const { getByText, queryByText } = render(<ShoppingListScreen />);

    expect(getByText('Útvonal és árak optimalizálása...')).toBeTruthy();
    expect(queryByText('Útvonal Optimalizálása')).toBeNull();
  });
});
