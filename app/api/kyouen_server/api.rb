# frozen_string_literal: true

module SharedParams
  params :pagination do
    optional :offset, type: Integer, default: 0
    optional :limit, type: Integer, default: 10, max_value: 100
  end
end

module KyouenServer
  class API < Grape::API
    version 'v1', using: :header, vendor: 'kyouen'
    format :json

    helpers SharedParams

    resource :stages do
      desc 'Return list of stages.'
      params do
        use :pagination
      end
      get '/' do
        KyouenPuzzle.fetch(params[:offset].to_i, params[:limit].to_i)
      end
    end
  end
end
