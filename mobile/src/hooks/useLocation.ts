import { useState, useCallback } from 'react';
import * as Location from 'expo-location';
import * as Sentry from '@sentry/react-native';
import { Alert } from 'react-native';
import { Coordinate } from '../services/api';

export function useLocation() {
  const [coords, setCoords] = useState<Coordinate | null>(null);

  const fetchLocation = useCallback(async (showErrorAlert: boolean = false): Promise<Coordinate | null> => {
    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status === 'granted') {
        const loc = await Location.getCurrentPositionAsync({});
        const newCoords = {
          latitude: loc.coords.latitude,
          longitude: loc.coords.longitude,
        };
        setCoords(newCoords);
        return newCoords;
      } else {
        if (showErrorAlert) {
          Alert.alert(
            'Helyadatok megtagadva',
            'A rendszer Budapest központjával tervez útvonalat.',
            [{ text: 'OK' }]
          );
        }
        return null;
      }
    } catch (error) {
      Sentry.captureException(error, { extra: { context: 'Hiba a helymeghatározás során:' } });
      if (showErrorAlert) {
        Alert.alert(
          'Helyadat hiba',
          'A rendszer Budapest központjával tervez útvonalat.',
          [{ text: 'OK' }]
        );
      }
      return null;
    }
  }, []);

  return { coords, setCoords, fetchLocation };
}
