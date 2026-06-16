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

export interface OptimizeResponse {
  route_plan: RoutePlan;
  total_cost: number;
}

export async function optimizeShoppingRoute(
  rawInput: string,
  latitude: number,
  longitude: number
): Promise<OptimizeResponse> {
  const url = 'http://localhost:8080/api/v1/optimize';
  
  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        raw_input: rawInput,
        user_latitude: latitude,
        user_longitude: longitude,
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data as OptimizeResponse;
  } catch (error) {
    console.error('Error optimizing shopping route:', error);
    throw error;
  }
}
