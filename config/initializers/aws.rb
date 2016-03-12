Aws.config.update({
  region: 'ap-northeast-1',
  credentials: Aws::Credentials.new(
    'akid',
    'secret'
  )
})
