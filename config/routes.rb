# frozen_string_literal: true
Rails.application.routes.draw do
  mount KyouenServer::API => '/'
  mount GrapeSwaggerRails::Engine => '/swagger'
end
