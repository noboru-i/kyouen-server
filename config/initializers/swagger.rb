# frozen_string_literal: true

GrapeSwaggerRails.options.url = '/swagger_doc.json'

GrapeSwaggerRails.options.before_action do
  GrapeSwaggerRails.options.app_url = request.protocol + request.host_with_port
  GrapeSwaggerRails.options.doc_expansion = 'full'

  GrapeSwaggerRails.options.api_key_name = 'X-Authorization'
  GrapeSwaggerRails.options.api_key_type = 'header'
end
