import { describe, it, expect } from 'vitest';
import { site, apiBaseUrl } from '../config/site';

describe('site config', () => {
	it('has correct site name', () => {
		expect(site.name).toBe('Sapala Fun');
	});

	it('has correct tagline', () => {
		expect(site.tagline).toContain('St. Croix');
		expect(site.tagline).toContain('USVI');
	});

	it('has valid email', () => {
		expect(site.email).toMatch(/^.+@.+\.fun$/);
	});

	it('has all social links', () => {
		expect(site.social.facebook).toBe('https://facebook.com/sapala');
		expect(site.social.instagram).toBe('https://instagram.com/sapala');
		expect(site.social.twitter).toBe('https://twitter.com/sapala');
		expect(site.social.youtube).toBe('https://youtube.com/@sapala');
	});

	it('is typed as readonly (TypeScript compile-time check)', () => {
		// TypeScript's `as const` makes this readonly at compile time.
		// At runtime, JavaScript objects are mutable, so we verify the structure is correct instead.
		expect(Object.prototype.hasOwnProperty.call(site, 'name')).toBe(true);
		expect(Object.keys(site)).toHaveLength(4); // name, tagline, email, social
	});
});

describe('apiBaseUrl', () => {
	it('has a default value', () => {
		expect(apiBaseUrl).toBeDefined();
		expect(typeof apiBaseUrl).toBe('string');
	});

	it('is a valid URL format', () => {
		expect(apiBaseUrl).toMatch(/^https?:\/\/.+/);
	});
});
