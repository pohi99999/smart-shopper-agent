import { useState, useEffect } from 'react';
import { Alert } from 'react-native';
import * as Location from 'expo-location';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Sentry from '@sentry/react-native';
import { optimizeShoppingRoute, OptimizeResponse, Coordinate } from '../services/api';
import { DEFAULT_LOCATION } from '../constants/location';

const ASYNC_STORAGE_KEY = '@last_shopping_result';

export function useShoppingOptimizer() {
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<OptimizeResponse | null>(null);
  const [coords, setCoords] = useState<Coordinate | null>(null);

  useEffect(() => {
    (async () => {
      try {
        // Load saved result from AsyncStorage
        const savedResultStr = await AsyncStorage.getItem(ASYNC_STORAGE_KEY);
        if (savedResultStr) {
          const savedResult = JSON.parse(savedResultStr);
          setResult(savedResult);
        }

        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status === 'granted') {
          const loc = await Location.getCurrentPositionAsync({});
          setCoords({
            latitude: loc.coords.latitude,
            longitude: loc.coords.longitude,
          });
        }
      } catch (error) {
        Sentry.captureException(error, { extra: { context: 'Hiba a kezdeti inicializáció során:' } });
      }
    })();
  }, []);

  const handleOptimize = async () => {
    if (!inputText.trim()) {
      Alert.alert('Hiba', 'Kérlek írd be a bevásárlólistádat!');
      return;
    }

    setLoading(true);
    setResult(null);

    let lat = DEFAULT_LOCATION.latitude;
    let lon = DEFAULT_LOCATION.longitude;

    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status === 'granted') {
        const loc = await Location.getCurrentPositionAsync({});
        lat = loc.coords.latitude;
        lon = loc.coords.longitude;
        setCoords({ latitude: lat, longitude: lon });
      } else {
        Alert.alert(
          'Helyadatok megtagadva',
          'A rendszer Budapest központjával tervez útvonalat.',
          [{ text: 'OK' }]
        );
      }
    } catch (error) {
      Sentry.captureException(error, { extra: { context: 'Hiba a helymeghatározás során:' } });
      Alert.alert(
        'Helyadat hiba',
        'A rendszer Budapest központjával tervez útvonalat.',
        [{ text: 'OK' }]
      );
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
