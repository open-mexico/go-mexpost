import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "go-mexpost",
  description: "Microservicio ultrarrápido para consulta de colonias, códigos postales y geocodificación inversa de México.",
  base: process.env.GITHUB_ACTIONS ? '/go-mexpost/' : '/',
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Inicio', link: '/' },
      { text: 'Configuración', link: '/configuracion' },
      { text: 'API', link: '/endpoints' }
    ],

    sidebar: [
      {
        text: 'Primeros pasos',
        items: [
          { text: 'Configuración e Instalación', link: '/configuracion' }
        ]
      },
      {
        text: 'Referencia de la API',
        items: [
          { text: 'Endpoints', link: '/endpoints' }
        ]
      },
      {
        text: 'Internals',
        items: [
          { text: 'Arquitectura Hexagonal', link: '/arquitectura' },
          { text: 'Base de Datos', link: '/base-de-datos' },
          { text: 'Geocodificación Inversa', link: '/geocodificacion' },
          { text: 'Pruebas', link: '/pruebas' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/open-mexico/go-mexpost' }
    ]
  }
})
