import * as Sentry from '@sentry/react-native';

export interface Coordinate {
  latitude: number;
  longitude: number;
}

export interface RouteStep {
  shop_name: string;
  items: string[];
}

export interface RoutePlan {
  steps: RouteStep[];
}

export interface OptimizeRequest {
  user_input: string;
  coords: Coordinate;
}

export interface OptimizeResponse {
  route_plan: RoutePlan;
  total_cost: number;
}

export interface APIError {
  error: string;
  code: number;
}

export async function optimizeShoppingRoute(
  rawInput: string,
  latitude: number,
  longitude: number
): Promise<OptimizeResponse> {
  const baseUrl = process.env.EXPO_PUBLIC_API_URL || 'https://localhost:8080';
  const url = `${baseUrl}/api/v1/optimize`;
  
  try {
    const requestBody: OptimizeRequest = {
      user_input: rawInput,
      coords: {
        latitude: latitude,
        longitude: longitude,
      }
    };

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      let errorMsg = `HTTP error! status: ${response.status}`;
      try {
        const errData = await response.json() as APIError;
        if (errData.error) {
          errorMsg = errData.error;
        }
      } catch (e) {
        // Fallback if not JSON
      }
      throw new Error(errorMsg);
    }

    const data = await response.json();
    return data as OptimizeResponse;
  } catch (error) {
    Sentry.captureException(error, { extra: { context: 'Error optimizing shopping route' } });
    throw error;
  }
}
