export const site = {
  name: 'Sapala Fun',
  tagline: 'Oceanfront Villas in St. Croix, USVI',
  email: 'info@sapala.fun',
  social: {
    facebook: 'https://facebook.com/sapala',
    instagram: 'https://instagram.com/sapala',
    twitter: 'https://twitter.com/sapala',
    youtube: 'https://youtube.com/@sapala',
  },
} as const;

export const apiBaseUrl = import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:3001';
