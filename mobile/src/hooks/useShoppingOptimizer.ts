import { useState, useEffect } from 'react';
import { Alert } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Sentry from '@sentry/react-native';
import { optimizeShoppingRoute, OptimizeResponse, Coordinate } from '../services/api';
import { DEFAULT_LOCATION } from '../constants/location';
import { useLocation } from './useLocation';

const ASYNC_STORAGE_KEY = '@last_shopping_result';

export function useShoppingOptimizer() {
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<OptimizeResponse | null>(null);
  const { coords, fetchLocation } = useLocation();

  useEffect(() => {
    (async () => {
      try {
        // Load saved result from AsyncStorage
        const savedResultStr = await AsyncStorage.getItem(ASYNC_STORAGE_KEY);
        if (savedResultStr) {
          const savedResult = JSON.parse(savedResultStr);
          setResult(savedResult);
        }

        await fetchLocation(false);
      } catch (error) {
        Sentry.captureException(error, { extra: { context: 'Hiba a kezdeti inicializáció során:' } });
      }
    })();
  }, [fetchLocation]);

  const handleOptimize = async () => {
    if (!inputText.trim()) {
      Alert.alert('Hiba', 'Kérlek írd be a bevásárlólistádat!');
      return;
    }

    setLoading(true);
    setResult(null);

    let lat = DEFAULT_LOCATION.latitude;
    let lon = DEFAULT_LOCATION.longitude;

    const newCoords = await fetchLocation(true);
    if (newCoords) {
      lat = newCoords.latitude;
      lon = newCoords.longitude;
    }

    try {
      const response = await optimizeShoppingRoute(inputText, lat, lon);
      setResult(response);
      // Save result to AsyncStorage
      await AsyncStorage.setItem(ASYNC_STORAGE_KEY, JSON.stringify(response));
    } catch (error) {
      Alert.alert(
        'Sikertelen optimalizálás',
        (error instanceof Error ? error.message : undefined) || 'Nem sikerült csatlakozni az optimalizáló szerverhez.'
      );
    } finally {
      setLoading(false);
    }
  };

  return {
    inputText,
    setInputText,
    loading,
    result,
    coords,
    handleOptimize,
  };
}
