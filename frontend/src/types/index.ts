export interface Property {
  id: string;
  slug: string;
  name: string;
  description: string;
  bedrooms: number;
  bathrooms: number;
  sleeps: string;
  nightlyRateMin: number;
  nightlyRateMax: number;
  imageUrl: string;
  amenities?: string[];
}
