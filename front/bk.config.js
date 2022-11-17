const { resolve } = require('path');
const replaceStaticUrlPlugin = require('./replace-static-url-plugin')
const isModeProduction = process.env.NODE_ENV === 'production';
const indexPath = isModeProduction ? './index.html' : './index-dev.html'
const env = require('./env')();
module.exports = {
  appConfig() {
    return {
      indexPath,
      mainPath: './src/main.ts',
      publicPath: env.publicPath,
      outputDir: env.outputDir,
      assetsDir: env.assetsDir,
      minChunkSize: 10000,
      // pages: {
      //   main: {
      //     entry: './src/main.ts',
      //     filename: 'index.html'
      //   },
      // },
      // needSplitChunks: false,
      css: {
        loaderOptions: {
          scss: {
            additionalData: '@import "./src/style/variables.scss";',
          },
        },
      },
      devServer : {
        host: env.DEV_HOST,
        port: 5000,
        historyApiFallback: true,
        proxy: {
          '/api/': {
            target: 'https://bk-metric-stag.woa.com/',
            ws: true,
            changeOrigin: true,
          },
          '/bkvision/': {
            target: 'https://bk-metric-stag.woa.com/',
            ws: false,
            changeOrigin: true,
          }
        }
      }
    }
  },
  configureWebpack(_webpackConfig) {
    webpackConfig = _webpackConfig;
    webpackConfig.plugins.push(
      new replaceStaticUrlPlugin(),
    )
    // webpackConfig.externals = {
    //   'axios':'axios',
    //   'dayjs':'dayjs',
    // }
    webpackConfig.resolve = {
      ...webpackConfig.resolve,
      symlinks: false,
      extensions: ['.js', '.vue', '.json', '.ts', '.tsx'],
      alias: {
        ...webpackConfig.resolve?.alias,
        // extensions: ['.js', '.jsx', '.ts', '.tsx'],
        '@': resolve(__dirname, './src'),
        '@static': resolve(__dirname, './static'),
        '@charts': resolve(__dirname, './src/plugins/charts'),
        '@datasource': resolve(__dirname, './src/plugins/datasource'),
        '@modules': resolve(__dirname, './src/store/modules'),
      },
    };
  },
};
