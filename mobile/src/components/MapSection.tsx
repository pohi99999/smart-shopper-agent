import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import MapView, { Marker } from 'react-native-maps';
import { OptimizeResponse, Coordinate } from '../services/api';

const SHOP_COORDINATES: { [key: string]: { latitude: number; longitude: number } } = {
  'Aldi': { latitude: 46.8451, longitude: 16.8455 },
  'Interspar': { latitude: 46.8413, longitude: 16.8521 },
};

interface MapSectionProps {
  coords: Coordinate | null;
  result: OptimizeResponse | null;
}

export function MapSection({ coords, result }: MapSectionProps) {
  return (
    <View style={styles.mapCard}>
      <Text style={styles.cardTitle}>Térkép</Text>
      <MapView
        style={styles.map}
        region={{
          latitude: coords ? coords.latitude : 47.4979,
          longitude: coords ? coords.longitude : 19.0402,
          latitudeDelta: 0.05,
          longitudeDelta: 0.05,
        }}
      >
        {coords && (
          <Marker
            coordinate={{
              latitude: coords.latitude,
              longitude: coords.longitude,
            }}
            title="Saját helyzet"
            pinColor="blue"
          />
        )}
        {result &&
          result.route_plan.steps.map((step, index) => {
            const shopCoords = SHOP_COORDINATES[step.shop_name];
            if (!shopCoords) return null;
            return (
              <Marker
                key={index}
                coordinate={shopCoords}
                title={`${index + 1}. állomás: ${step.shop_name}`}
                description={step.items.join(', ')}
                pinColor="red"
              />
            );
          })}
      </MapView>
    </View>
  );
}

const styles = StyleSheet.create({
  mapCard: {
    backgroundColor: '#FFFFFF',
    borderRadius: 16,
    padding: 18,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.08,
    shadowRadius: 12,
    elevation: 3,
    marginBottom: 24,
    overflow: 'hidden',
  },
  cardTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1C1C1E',
    marginBottom: 12,
  },
  map: {
    width: '100%',
    height: 250,
    borderRadius: 12,
  },
});
